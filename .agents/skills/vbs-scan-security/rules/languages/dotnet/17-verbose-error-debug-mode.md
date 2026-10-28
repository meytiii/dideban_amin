---
id: VERBOSE-ERROR-DEBUG-MODE
severity_max: HIGH
applies_to: dotnet
overrides: generic/17-verbose-error-debug-mode.md
---

# Verbose Error / Debug Mode — .NET 10 Specialization

> Override cho rule chung `rules/generic/17-verbose-error-debug-mode.md`. Áp dụng khi `primary_language: dotnet`.

## Intent (.NET-specific)

ASP.NET Core có **Developer Exception Page** (`app.UseDeveloperExceptionPage()`): khi request lỗi, trang này trả stack trace đầy đủ, source snippet, query string, cookie, header và route data cho client. Đây là analog của Flask debug / Laravel Ignition. Mặc định `WebApplication.CreateBuilder` chỉ bật nó khi `ASPNETCORE_ENVIRONMENT=Development`; rủi ro xuất hiện khi dev bật vô điều kiện, hoặc khi môi trường production chạy với `Development`.

## Khi nào HIGH

- `app.UseDeveloperExceptionPage()` gọi vô điều kiện (không bọc trong `if (app.Environment.IsDevelopment())`).
- `ASPNETCORE_ENVIRONMENT=Development` / `DOTNET_ENVIRONMENT=Development` trong file deploy production (Dockerfile, `docker-compose.yml`, `launchSettings.json` được deploy, manifest k8s).
- `app.Environment.EnvironmentName = "Development"` hard-code.
- `<customErrors mode="Off">` trong `web.config` (ASP.NET Framework / IIS).
- Exception handler tự viết trả `ex.ToString()`, `ex.StackTrace` hoặc `ex.InnerException` trong response body.
- `ProblemDetails` được nhét `Detail = ex.ToString()` / `Extensions["stackTrace"]` ở mọi môi trường.

## Khi nào MEDIUM (giảm cấp)

- `UseDeveloperExceptionPage()` chỉ bật qua config flag mà default là tắt.
- `launchSettings.json` chứa `Development` nhưng chỉ dùng local (không được copy vào image/deploy).
- `<compilation debug="true">` trong `web.config` (không tự lộ lỗi ra client, chủ yếu ảnh hưởng hiệu năng/symbol; HIGH nếu đi kèm `customErrors mode="Off"`).
- Stack trace chỉ ghi vào log (`logger.LogError(ex, ...)`), không trả về response.

## Cách reasoning

1. Grep sink: `UseDeveloperExceptionPage`, `EnvironmentName`, `ASPNETCORE_ENVIRONMENT`, `customErrors`, `ex.StackTrace`, `ex.ToString()`.
2. Read `Program.cs` / `Startup.Configure` và xác định:
   - Lời gọi có nằm trong `if (env.IsDevelopment())` không?
   - Nhánh `else` có `UseExceptionHandler("/error")` (+ `UseHsts`) không?
3. Với env var: xác định file có phải cấu hình deploy production hay chỉ là local dev.
4. Với custom handler: kiểm tra response body có chứa `ex.Message`/`StackTrace` cho mọi môi trường hay đã gate theo `IsDevelopment()`.
5. Chỉ report khi debug/verbose path có thể tới client ở production.

## Search patterns (gợi ý — KHÔNG chạy literal, dùng Grep tool)

```
UseDeveloperExceptionPage\s*\(
EnvironmentName\s*=\s*["']Development["']
(ASPNETCORE|DOTNET)_ENVIRONMENT\s*[=:]\s*["']?Development
<customErrors\s+mode="Off"
<compilation[^>]+debug="true"
(StackTrace|ex\.ToString\(\)|\.InnerException)   # trong response body
Detail\s*=\s*(ex|e|exception)\.(ToString|StackTrace)
```

## Examples

### HIGH — flag

```csharp
// Program.cs — bật vô điều kiện
var app = builder.Build();
app.UseDeveloperExceptionPage();
app.MapControllers();
app.Run();
```

```dockerfile
# Dockerfile dùng để deploy production
ENV ASPNETCORE_ENVIRONMENT=Development
ENTRYPOINT ["dotnet", "Shop.dll"]
```

```csharp
// Exception handler trả stack trace cho client
app.UseExceptionHandler(errorApp => errorApp.Run(async ctx =>
{
    var ex = ctx.Features.Get<IExceptionHandlerFeature>()?.Error;
    ctx.Response.StatusCode = 500;
    await ctx.Response.WriteAsync(ex?.ToString() ?? "error");
}));
```

### NOT flag

```csharp
// Chỉ bật ở Development, production dùng handler chung
if (app.Environment.IsDevelopment())
{
    app.UseDeveloperExceptionPage();
}
else
{
    app.UseExceptionHandler("/error");
    app.UseHsts();
}
```

```csharp
// Log nội bộ, trả message generic
app.UseExceptionHandler(errorApp => errorApp.Run(async ctx =>
{
    var ex = ctx.Features.Get<IExceptionHandlerFeature>()?.Error;
    logger.LogError(ex, "Unhandled exception");
    ctx.Response.StatusCode = 500;
    await ctx.Response.WriteAsJsonAsync(new { error = "Internal server error" });
}));
```

## Fix recommendation

1. Bọc `UseDeveloperExceptionPage()` trong `if (app.Environment.IsDevelopment())`; production dùng `UseExceptionHandler("/error")`.
2. Đặt `ASPNETCORE_ENVIRONMENT=Production` trong Dockerfile/manifest deploy; giữ `Development` chỉ ở `launchSettings.json` local.
3. Log exception đầy đủ phía server (`ILogger`, Sentry/App Insights), trả `ProblemDetails` generic kèm `traceId` cho client.
4. Với ASP.NET Framework: `<customErrors mode="RemoteOnly">` hoặc `On`, `<compilation debug="false">`.

## Cross-references

- `01-hardcoded-secret`: stack trace / env dump có thể lộ connection string và secret.
- `02-sql-injection`: SQL error message lộ schema, giúp khai thác dễ hơn.
- `14-jwt-none-algorithm`: thông báo lỗi validate token có thể lộ chi tiết cấu hình key.
