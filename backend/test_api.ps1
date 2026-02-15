try {
    $body = @{ email = "testuser99@example.com"; password = "test123" } | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "http://localhost:8080/register" -Method POST -ContentType "application/json" -Body $body
    Write-Host "SUCCESS:" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "FAILED:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host $reader.ReadToEnd()
    }
}
