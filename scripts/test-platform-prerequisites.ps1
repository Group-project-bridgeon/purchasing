#Requires -Version 7.4
[CmdletBinding()]
param([string]$GoExe='go',[string]$AuthSource,[switch]$SkipImage)
$ErrorActionPreference='Stop'
$PSNativeCommandUseErrorActionPreference=$false
$root=Split-Path $PSScriptRoot -Parent
$run='purchasing-prerequisites-'+[guid]::NewGuid().ToString('N')
$runID=$run.Substring('purchasing-prerequisites-'.Length)
$temporary=Join-Path $root ('.'+$run)
$database='purchasing_test_'+$runID
$refusalDatabase='purchasing_refusal_'+$runID
$authDatabase='auth_purchasing_'+$runID
$allowedDatabases=@('postgres',$database,$refusalDatabase,$authDatabase)
$containerID=$null
$appID=$null
$resourceWorkStarted=$false
$imageWorkStarted=$false
$image='erp/purchasing-prerequisites:'+$runID
$password=[Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(32)).ToLowerInvariant()
$names=@('POSTGRES_PASSWORD','PGPASSWORD','DATABASE_URL','JWT_PUBLIC_KEY_PATH','PURCHASING_TEST_DATABASE_URL','PURCHASING_TEST_RUN_ID','GOMAXPROCS','GOTOOLCHAIN','GOTMPDIR')
$previous=@{}
foreach($name in $names){$previous[$name]=[Environment]::GetEnvironmentVariable($name,'Process')}
$dockerExe=if(Get-Command 'docker.exe' -ErrorAction SilentlyContinue){'docker.exe'}else{'docker'}

function InvokeFixtureDocker([string[]]$Arguments){
 $out=& $dockerExe @Arguments 2>&1
 if($LASTEXITCODE -ne 0){throw "Owned Docker operation failed: $($Arguments[0]); raw output withheld"}
 return $out
}
function Owned([string]$ID){
 if($ID -notmatch '^[a-f0-9]{64}$'){throw 'Invalid owned-container identity'}
 $label=[string](InvokeFixtureDocker @('inspect','--format','{{index .Config.Labels "purchasing.prerequisites.fixture"}}',$ID))
 if($label.Trim() -ne $run){throw 'Refusing operation on a container not owned by this run'}
}
function SQL([string]$DB,[string]$Query,[string]$User='purchasing_fixture'){
 Owned $containerID
 if($DB -notin $allowedDatabases){throw 'Refusing an unallowlisted database'}
 $out=InvokeFixtureDocker @('exec','-e','PGPASSWORD','-e','PGOPTIONS=-c statement_timeout=5000 -c lock_timeout=2000',$containerID,'psql','-X','-t','-A','-v','ON_ERROR_STOP=1','-U',$User,'-d',$DB,'-c',$Query)
 return ([string]::Join("`n",[string[]]@($out))).Trim()
}
function AssertSQL([string]$DB,[string]$Query,[string]$Expected,[string]$Label,[string]$User='purchasing_fixture'){
 $actual=SQL $DB $Query $User
 if($actual -ne $Expected){if($actual -match '^(?:[0-9]+(?:\|(?:true|false))?|[tf])$'){throw "Fixture assertion failed: $Label; observed safe count/version $actual"};throw "Fixture assertion failed: $Label; result withheld"};Write-Host "PASS: $Label"
}
function Migrate([string]$Source,[string]$DB,[string]$User,[string[]]$Operation,[switch]$ExpectFailure){
 Owned $containerID
 if($DB -notin $allowedDatabases){throw 'Refusing an unallowlisted migration target'}
 $migrationArgs=@('run','--rm','--label',"purchasing.prerequisites.fixture=$run",'--network',"container:$containerID",'-e','PGPASSWORD','--mount',"type=bind,src=$Source,dst=/migrations,readonly",'migrate/migrate:v4.18.3','-path=/migrations','-lock-timeout=15',"-database=postgres://${User}@127.0.0.1:5432/${DB}?sslmode=disable&connect_timeout=5&statement_timeout=10000")+$Operation
 $out=& $dockerExe @migrationArgs 2>&1
 $code=$LASTEXITCODE
 if($ExpectFailure){if($code -eq 0){throw 'Populated authority downgrade unexpectedly succeeded'}}elseif($code -ne 0){throw 'Disposable migration failed; raw output withheld'}
}
function GoCheck([string[]]$Arguments){
 $out=& $GoExe -C $root @Arguments 2>&1
 $code=$LASTEXITCODE
 foreach($line in $out){Write-Host (([string]$line).Replace($password,'[fixture-secret-redacted]'))}
 if($code -ne 0){throw 'Purchasing Go check failed; owned cleanup will run'}
}
function Snapshot{
 $result=@{}
 foreach($id in @(InvokeFixtureDocker @('ps','-aq','--no-trunc'))){if($id){$result[$id]=([string](InvokeFixtureDocker @('inspect','--format','{{.Id}}|{{.State.StartedAt}}|{{.State.Running}}|{{.RestartCount}}',$id))).Trim()}}
 return $result
}
function HTTPStatus([string]$Path){
 Owned $appID
 $out=& $dockerExe exec $appID wget -T 3 -S -O /dev/null "http://127.0.0.1:8080$Path" 2>&1
 $code=$LASTEXITCODE
 $match=[regex]::Match(([string]::Join("`n",[string[]]@($out))),'HTTP/\S+\s+(\d{3})')
 if(!$match.Success){throw 'Owned image HTTP status unavailable; raw response withheld'}
 return [int]$match.Groups[1].Value
}

$before=$null
$cleanupFailures=[Collections.Generic.List[string]]::new()
try{
 InvokeFixtureDocker @('version','--format','{{.Server.Version}}')|Out-Null
 if($AuthSource){$AuthSource=(Resolve-Path -LiteralPath $AuthSource).Path;if(!(Test-Path -LiteralPath (Join-Path $AuthSource 'db/migrations/000013_purchasing_permissions.up.sql'))){throw 'Explicit Auth source lacks the reviewed Purchasing permission migration'}}
 $env:GOTOOLCHAIN='go1.27.3';$env:GOMAXPROCS='2'
 $compiler=& $GoExe version 2>&1
 if($LASTEXITCODE -ne 0 -or [string]$compiler -notmatch 'go1\.27\.3 '){throw 'Verified Go 1.27.3 is required'}
 Write-Host "Checking existing container state before creating the owned fixture"
 $before=Snapshot
 New-Item -ItemType Directory -Path $temporary|Out-Null
 [IO.File]::WriteAllText((Join-Path $temporary 'owner.txt'),$run)
 New-Item -ItemType Directory -Path (Join-Path $temporary 'go-tmp')|Out-Null
 $env:GOTMPDIR=Join-Path $temporary 'go-tmp'
 $env:POSTGRES_PASSWORD=$password;$env:PGPASSWORD=$password
 $resourceWorkStarted=$true
 Write-Host "Starting uniquely labelled tmpfs PostgreSQL fixture"
 $containerID=([string](InvokeFixtureDocker @('run','--detach','--name',$run,'--label',"purchasing.prerequisites.fixture=$run",'--tmpfs','/var/lib/postgresql/data:rw,nosuid,nodev','--publish','127.0.0.1::5432','-e','POSTGRES_USER=prerequisite_admin','-e','POSTGRES_PASSWORD','-e','POSTGRES_DB=postgres','-e','POSTGRES_INITDB_ARGS=--auth-host=scram-sha-256','postgres:17-alpine'))).Trim()
 Owned $containerID
 $ready=$false
 for($attempt=0;$attempt -lt 60;$attempt++){
  $null=& $dockerExe exec $containerID pg_isready -U prerequisite_admin -d postgres 2>$null
  if($LASTEXITCODE -eq 0){$ready=$true;break};Start-Sleep -Milliseconds 500
 }
 if(!$ready){throw 'Owned PostgreSQL readiness timed out'}
 SQL 'postgres' "CREATE ROLE purchasing_fixture LOGIN PASSWORD '$password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION" 'prerequisite_admin'|Out-Null
 foreach($db in @($database,$refusalDatabase)){
  SQL 'postgres' "CREATE DATABASE $db OWNER purchasing_fixture" 'prerequisite_admin'|Out-Null
  SQL 'postgres' "REVOKE CONNECT ON DATABASE $db FROM PUBLIC; GRANT CONNECT ON DATABASE $db TO purchasing_fixture" 'prerequisite_admin'|Out-Null
 }
 AssertSQL $database "SELECT NOT rolsuper AND NOT rolcreatedb AND NOT rolcreaterole FROM pg_roles WHERE rolname=current_user" 't' 'service-owned non-admin migration/runtime role'
 $migrations=Join-Path $root 'db/migrations'
 Migrate $migrations $database 'purchasing_fixture' @('up','8')
 SQL $database "INSERT INTO purchase_orders(po_number,vendor_id,status,currency) VALUES('LEGACY-UNKNOWN','30000000-0000-4000-8000-000000000001','draft','USD')"|Out-Null
 Migrate $migrations $database 'purchasing_fixture' @('up')
 AssertSQL $database "SELECT version::text||'|'||dirty::text FROM schema_migrations" '9|false' 'fresh golang-migrate schema 9'
 Migrate $migrations $database 'purchasing_fixture' @('down','1')
 AssertSQL $database "SELECT count(*) FROM purchase_orders WHERE po_number='LEGACY-UNKNOWN'" '1' 'empty authority downgrade retains legacy business facts'
 Migrate $migrations $database 'purchasing_fixture' @('up')
 AssertSQL $database "SELECT count(*) FROM purchase_orders WHERE po_number='LEGACY-UNKNOWN' AND created_by IS NULL" '1' 'upgrade never guesses historical creator'
 SQL $database "CREATE TABLE purchasing_fixture_ownership(run_id TEXT PRIMARY KEY);INSERT INTO purchasing_fixture_ownership VALUES('$runID')"|Out-Null
 Migrate $migrations $refusalDatabase 'purchasing_fixture' @('up')
 SQL $refusalDatabase "INSERT INTO purchase_orders(po_number,vendor_id,status,currency,created_by) VALUES('REFUSAL-FACT','30000000-0000-4000-8000-000000000001','draft','USD','30000000-0000-4000-8000-000000000002')"|Out-Null
 Migrate $migrations $refusalDatabase 'purchasing_fixture' @('down','1') -ExpectFailure
 AssertSQL $refusalDatabase "SELECT count(*) FROM purchase_orders WHERE po_number='REFUSAL-FACT' AND created_by='30000000-0000-4000-8000-000000000002'" '1' 'populated downgrade refuses without deleting authority facts'
 AssertSQL $refusalDatabase "SELECT version::text||'|'||dirty::text FROM schema_migrations" '8|true' 'refused downgrade leaves explicit dirty marker; no force/baseline'
 if($AuthSource){
  SQL 'postgres' "CREATE ROLE auth_purchasing_fixture LOGIN PASSWORD '$password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION" 'prerequisite_admin'|Out-Null
  SQL 'postgres' "CREATE DATABASE $authDatabase OWNER auth_purchasing_fixture" 'prerequisite_admin'|Out-Null
  SQL 'postgres' "REVOKE CONNECT ON DATABASE $authDatabase FROM PUBLIC;GRANT CONNECT ON DATABASE $authDatabase TO auth_purchasing_fixture" 'prerequisite_admin'|Out-Null
  Migrate (Join-Path $AuthSource 'db/migrations') $authDatabase 'auth_purchasing_fixture' @('up')
  AssertSQL $authDatabase "SELECT count(*) FROM identity_permissions WHERE resource LIKE 'purchasing.%'" '41' 'Auth adds exactly 41 Purchasing definitions' 'auth_purchasing_fixture'
  AssertSQL $authDatabase "SELECT count(*) FROM identity_role_permissions rp JOIN identity_permissions p ON p.id=rp.permission_id WHERE p.resource LIKE 'purchasing.%'" '0' 'Auth adds no automatic Purchasing grants' 'auth_purchasing_fixture'
  AssertSQL $authDatabase "SELECT count(*) FROM identity_permissions WHERE resource='purchasing.outbox'" '0' 'Outbox authority remains absent' 'auth_purchasing_fixture'
 }else{Write-Host 'SKIP: Auth definition persistence (no explicit -AuthSource; Purchasing CI remains independent of sibling checkouts)'}
 $port=([string](InvokeFixtureDocker @('port',$containerID,'5432/tcp'))).Trim().Split(':')[-1]
 if($port -notmatch '^\d+$' -or [int]$port -in @(5432,5433)){throw 'Owned loopback PostgreSQL port is unsafe'}
 $env:PURCHASING_TEST_RUN_ID=$runID
 $env:PURCHASING_TEST_DATABASE_URL="postgres://purchasing_fixture:${password}@127.0.0.1:${port}/${database}?sslmode=disable&statement_timeout=5000&lock_timeout=2000&idle_in_transaction_session_timeout=10000"
 GoCheck @('test','-p=1','-count=1','-timeout=240s','./...')
 if(!$SkipImage){
  $imageWorkStarted=$true
  InvokeFixtureDocker @('build','--label',"purchasing.prerequisites.fixture=$run",'-t',$image,$root)|Out-Null
  $rsa=[Security.Cryptography.RSA]::Create(2048)
  try{[IO.File]::WriteAllText((Join-Path $temporary 'public.pem'),$rsa.ExportSubjectPublicKeyInfoPem())}finally{$rsa.Dispose()}
  $env:DATABASE_URL="postgres://purchasing_fixture@127.0.0.1:5432/${database}?sslmode=disable"
  $appID=([string](InvokeFixtureDocker @('run','--detach','--name',"$run-app",'--label',"purchasing.prerequisites.fixture=$run",'--network',"container:$containerID",'--read-only','--tmpfs','/tmp:rw,nosuid,nodev','--security-opt','no-new-privileges:true','-e','DATABASE_URL','-e','PGPASSWORD','-e','JWT_PUBLIC_KEY_PATH=/run/fixture/public.pem','-e','APP_ENV=production','-e','LOG_FORMAT=json','--mount',"type=bind,src=$temporary/public.pem,dst=/run/fixture/public.pem,readonly",$image))).Trim()
  Owned $appID
  $imageReady=$false
  for($attempt=0;$attempt -lt 60;$attempt++){try{if((HTTPStatus '/ready') -eq 200){$imageReady=$true;break}}catch{};Start-Sleep -Milliseconds 500}
  if(!$imageReady){throw 'Owned application readiness timed out'}
  if(([string](InvokeFixtureDocker @('exec',$appID,'id','-u'))).Trim() -ne '10001'){throw 'Application did not run as configured non-root UID'}
  InvokeFixtureDocker @('exec',$appID,'test','-r','/run/fixture/public.pem')|Out-Null
  if((HTTPStatus '/api/v1/purchase-orders') -ne 401){throw 'Anonymous image business request did not return 401'}
  Write-Host 'PASS: built image, non-root UID, read-only public key, readiness and anonymous 401'
  Owned $containerID;InvokeFixtureDocker @('stop',$containerID)|Out-Null
  if((HTTPStatus '/ready') -ne 503 -or (HTTPStatus '/health') -ne 200){throw 'Actual dependency outage probe contract failed'}
  Write-Host 'PASS: actual owned PostgreSQL outage gives readiness 503 and liveness 200'
 }else{Write-Host 'SKIP: image/runtime checks explicitly requested with -SkipImage'}
}finally{
 if($resourceWorkStarted){try{
  $ownedIDs=@(InvokeFixtureDocker @('ps','-aq','--no-trunc','--filter',"label=purchasing.prerequisites.fixture=$run")) | Sort-Object {if($_ -eq $containerID){1}else{0}}
  foreach($id in $ownedIDs){if($id){Owned $id;InvokeFixtureDocker @('rm','--force',$id)|Out-Null}}
 }catch{$cleanupFailures.Add('Owned container cleanup could not be confirmed')}}
 if($imageWorkStarted){try{
  $label=& $dockerExe image inspect --format '{{index .Config.Labels "purchasing.prerequisites.fixture"}}' $image 2>$null
  if($LASTEXITCODE -eq 0){if(([string]$label).Trim() -ne $run){throw 'Image label mismatch'};InvokeFixtureDocker @('image','rm',$image)|Out-Null}
 }catch{$cleanupFailures.Add('Owned image cleanup could not be confirmed')}}
 if(Test-Path -LiteralPath $temporary){
  $resolved=[IO.Path]::GetFullPath($temporary);$prefix=[IO.Path]::GetFullPath($root)+[IO.Path]::DirectorySeparatorChar
  if(!$resolved.StartsWith($prefix,[StringComparison]::OrdinalIgnoreCase) -or (Split-Path $resolved -Leaf) -ne ".$run" -or (Get-Content -LiteralPath (Join-Path $resolved 'owner.txt') -Raw) -ne $run){$cleanupFailures.Add('Temporary ownership guard refused cleanup')}
  else{Remove-Item -LiteralPath $resolved -Recurse -Force}
 }
 foreach($name in $names){[Environment]::SetEnvironmentVariable($name,$previous[$name],'Process')}
 if($null -ne $before){try{$after=Snapshot;foreach($id in $before.Keys){if(!$after.ContainsKey($id) -or $after[$id] -ne $before[$id]){throw 'Pre-existing container state changed'}};Write-Host "PASS: $($before.Count) pre-existing containers preserved"}catch{$cleanupFailures.Add('Pre-existing container preservation could not be confirmed')}}
 if($cleanupFailures.Count){throw ([string]::Join('; ',$cleanupFailures))}
 Write-Host 'PASS: run-owned cleanup and process environment restoration'
}
