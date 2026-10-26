$ErrorActionPreference = "Stop"

$backupDir = if ($env:HERASAT_BACKUP_DIR) { $env:HERASAT_BACKUP_DIR } else { ".\backups" }
if (-not (Test-Path $backupDir)) {
    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
}

$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$filename = "backup_herasat_$timestamp.sql"
$gzFilename = "$filename.gz"
$targetPath = Join-Path $backupDir $filename
$targetGzPath = Join-Path $backupDir $gzFilename

$containerName = "herasat_postgres"
$dbUser = "herasat_user"
$dbName = "herasat_db"

if (Get-Command pg_dump -ErrorAction SilentlyContinue) {
    if ($env:HERASAT_DB_CONN) {
        & pg_dump --dbname $env:HERASAT_DB_CONN --no-owner --no-acl --clean --if-exists -f $targetPath
    } else {
        & pg_dump -h 127.0.0.1 -p 5432 -U $dbUser --no-owner --no-acl --clean --if-exists -d $dbName -f $targetPath
    }
} elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    & docker exec $containerName pg_dump -U $dbUser --no-owner --no-acl --clean --if-exists $dbName > $targetPath
} else {
    Write-Error "Neither pg_dump nor docker command found."
    exit 1
}

if (Test-Path $targetPath) {
    $srcStream = [System.IO.File]::OpenRead((Resolve-Path $targetPath))
    $destStream = [System.IO.File]::Create((Resolve-Path -Path $backupDir | Join-Path -ChildPath $gzFilename))
    $gzipStream = New-Object System.IO.Compression.GZipStream($destStream, [System.IO.Compression.CompressionMode]::Compress)
    $srcStream.CopyTo($gzipStream)
    $gzipStream.Close()
    $destStream.Close()
    $srcStream.Close()
    Remove-Item -Path $targetPath -Force
    Write-Output "BACKUP_SUCCESS: $targetGzPath"
}
