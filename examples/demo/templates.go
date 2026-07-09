package main

const indexTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Payment Gateway Go SDK Demo</title>
  <style>` + demoCSS + `</style>
</head>
<body>
<header class="topbar">
  <div>
    <h1>Payment Gateway Go SDK Demo</h1>
    <p>商户号 {{.MerchantNo}}</p>
  </div>
</header>
<main class="layout">
  {{range $group, $apis := .Groups}}
  <section class="band">
    <h2>{{$group}}</h2>
    <div class="api-grid">
      {{range $apis}}
      <a class="api-row" href="/demo/api/{{.Code}}">
        <span>
          <strong>{{.Name}}</strong>
          <small>{{.Description}}</small>
        </span>
        <em>{{.Method}}</em>
      </a>
      {{end}}
    </div>
  </section>
  {{end}}
</main>
</body>
</html>`

const apiTemplate = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.API.Name}} - Go SDK Demo</title>
  <style>` + demoCSS + `</style>
</head>
<body>
<header class="topbar">
  <div>
    <a href="/demo/apis" class="back">← API 列表</a>
    <h1>{{.API.Name}}</h1>
    <p>{{.API.Description}}</p>
  </div>
  <code>{{.API.Method}} {{.API.Path}}</code>
</header>
<main class="workbench">
  <form method="post" class="panel">
    <div class="field-list">
      {{range .API.Fields}}
      <label class="field" data-field="{{.Name}}">
        <span>{{.Label}}{{if .Required}} <b>*</b>{{end}}</span>
        <small>{{.Description}}</small>
        {{if eq .Name "merchantNo"}}
          <input name="{{.Name}}" value="{{$.MerchantNo}}" readonly>
        {{else if eq .Type "json"}}
          <textarea name="{{.Name}}" rows="8">{{fieldValue . $.FormValues $.MerchantNo}}</textarea>
        {{else if or (eq .Type "select") (eq .Type "array")}}
          <select name="{{.Name}}">
            {{$field := .}}{{range .Options}}<option value="{{.}}" {{selected . $field.Name $.FormValues $field.Default}}>{{.}}</option>{{end}}
          </select>
        {{else}}
          <input name="{{.Name}}" value="{{fieldValue . $.FormValues $.MerchantNo}}">
        {{end}}
      </label>
      {{end}}
    </div>
    <button type="submit">{{.API.Action}}</button>
  </form>
  <section class="panel result">
    {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
    <h2>请求参数</h2>
    <pre>{{.Request}}</pre>
    <h2>响应参数</h2>
    <pre>{{if .Response}}{{.Response}}{{else}}提交后展示网关响应{{end}}</pre>
    <h2>字段说明</h2>
    <dl>
      {{range .API.Response}}
      <dt>{{.Name}}</dt><dd>{{.Description}}</dd>
      {{end}}
    </dl>
  </section>
</main>
<script>
const methodData = {{.MethodsJSON}};
function syncMethodData() {
  const method = document.querySelector('[name="paymentMethod"]');
  const data = document.querySelector('[name="paymentMethodData"]');
  if (method && data && methodData[method.value]) {
    data.value = JSON.stringify(methodData[method.value], null, 2);
  }
}
function syncCustomerMode() {
  const mode = document.querySelector('[name="customerMode"]');
  const customerId = document.querySelector('[data-field="customerId"]');
  const customer = document.querySelector('[data-field="customer"]');
  if (!mode || !customerId || !customer) return;
  customerId.style.display = mode.value === 'customerId' ? 'grid' : 'none';
  customer.style.display = mode.value === 'customer' ? 'grid' : 'none';
}
document.addEventListener('change', event => {
  if (event.target.name === 'paymentMethod') syncMethodData();
  if (event.target.name === 'customerMode') syncCustomerMode();
});
syncCustomerMode();
</script>
</body>
</html>`

const demoCSS = `
:root{--ink:#17211b;--muted:#5f6f68;--line:#d9e0dc;--panel:#f9fbfa;--accent:#0f766e;--warn:#b42318;--code:#101828}
*{box-sizing:border-box}
body{margin:0;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:var(--ink);background:#eef3f1}
.topbar{display:flex;align-items:end;justify-content:space-between;gap:24px;padding:28px 36px;background:#ffffff;border-bottom:1px solid var(--line)}
.topbar h1{margin:6px 0 4px;font-size:28px;line-height:1.15;font-weight:750;letter-spacing:0}
.topbar p{margin:0;color:var(--muted)}
.topbar code{padding:8px 10px;border:1px solid var(--line);background:#f4f7f6;color:var(--code);border-radius:6px;white-space:nowrap}
.back{color:var(--accent);text-decoration:none;font-weight:650}
.layout{max-width:1180px;margin:0 auto;padding:28px 24px 56px}
.band{margin-bottom:28px}
.band h2{font-size:18px;margin:0 0 12px}
.api-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));gap:10px}
.api-row{display:flex;justify-content:space-between;gap:12px;padding:14px 16px;background:#fff;border:1px solid var(--line);border-radius:8px;color:inherit;text-decoration:none}
.api-row:hover{border-color:var(--accent)}
.api-row strong{display:block;font-size:15px}
.api-row small{display:block;margin-top:5px;color:var(--muted);line-height:1.45}
.api-row em{font-style:normal;color:var(--accent);font-weight:700;font-size:12px}
.workbench{display:grid;grid-template-columns:minmax(360px,520px) minmax(0,1fr);gap:18px;padding:24px 28px 56px}
.panel{background:#fff;border:1px solid var(--line);border-radius:8px;padding:18px}
.field-list{display:grid;gap:14px}
.field{display:grid;gap:6px}
.field span{font-weight:700;font-size:14px}
.field b{color:var(--warn)}
.field small{color:var(--muted);line-height:1.35}
input,select,textarea{width:100%;border:1px solid var(--line);border-radius:6px;padding:10px 11px;font:14px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace;background:var(--panel);color:var(--ink)}
textarea{resize:vertical}
button{margin-top:18px;width:100%;border:0;border-radius:6px;padding:12px 14px;background:var(--accent);color:#fff;font-weight:800;cursor:pointer}
button:hover{filter:brightness(.95)}
.result h2{font-size:16px;margin:0 0 10px}
.result h2:not(:first-child){margin-top:18px}
pre{margin:0;overflow:auto;white-space:pre-wrap;background:#0b1411;color:#d9f5e9;border-radius:8px;padding:14px;min-height:80px;font:13px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace}
.error{border:1px solid #f3b5ae;background:#fff4f2;color:var(--warn);padding:10px 12px;border-radius:6px;margin-bottom:14px}
dl{display:grid;grid-template-columns:120px 1fr;gap:8px 14px;margin:0}
dt{font-weight:800}
dd{margin:0;color:var(--muted)}
@media(max-width:900px){.topbar{display:block;padding:22px}.workbench{grid-template-columns:1fr;padding:18px}.topbar code{display:block;margin-top:12px;white-space:normal}.layout{padding:18px}}
`
