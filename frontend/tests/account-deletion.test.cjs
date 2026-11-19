const test = require('node:test'),
  assert = require('node:assert/strict');
const fs = require('node:fs'),
  path = require('node:path'),
  vm = require('node:vm'),
  ts = require('typescript'),
  rx = require('rxjs');

class HttpErrorResponse extends Error {
  constructor(v) {
    super('HTTP ' + v.status);
    Object.assign(this, v);
  }
}
function transpile(file) {
  return ts.transpileModule(fs.readFileSync(path.join(__dirname, '..', file), 'utf8'), {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2022,
      experimentalDecorators: true,
    },
  }).outputText;
}
function load(file, requireStub) {
  const module = { exports: {} };
  vm.runInNewContext(transpile(file), {
    module,
    exports: module.exports,
    require: requireStub,
  });
  return module.exports;
}
const account = () =>
  load('src/app/core/services/account.service.ts', (name) =>
    name === 'rxjs'
      ? rx
      : name === '@angular/core'
        ? { Injectable: () => () => {} }
        : { HttpErrorResponse },
  );

test('only a finished deletion is reported as deleted', () => {
  const { deletionOutcome } = account();
  assert.equal(deletionOutcome(200, { data: { status: 'deleted' } }), 'deleted');
  assert.equal(deletionOutcome(202, { data: { status: 'pending' } }), 'pending');
});

test('unexpected responses never count as success', () => {
  const { deletionOutcome } = account();
  for (const [status, body] of [
    [200, { data: { status: 'pending' } }],
    [202, { data: { status: 'deleted' } }],
    [200, {}],
    [200, null],
    [204, null],
    [202, { data: {} }],
  ])
    assert.throws(() => deletionOutcome(status, body), (e) => e.status === 502);
});

test('failures state that the account was not deleted or the result is unknown', () => {
  const { deletionError } = account();
  const text = (status) => deletionError(new HttpErrorResponse({ status }));
  assert.match(text(403), /Неверный пароль.*не удалён/);
  assert.match(text(401), /Войдите снова/);
  assert.match(text(404), /уже удалён/);
  assert.match(text(429), /Слишком много/);
  assert.match(text(0), /не подтверждено/);
  assert.match(text(503), /не подтверждено/);
  const timeout = new Error('timeout');
  timeout.name = 'TimeoutError';
  assert.match(deletionError(timeout), /не подтверждено/);
  for (const message of [text(403), text(0), text(500), deletionError(null)])
    assert.doesNotMatch(message, /Аккаунт удалён\./);
});

test('service sends DELETE /api/v1/account with the password and maps the status', async () => {
  const { AccountService } = account();
  let request;
  const reply = (status, body) => (method, url, options) => {
    request = { method, url, options };
    return rx.of({ status, body });
  };
  const http = { request: reply(200, { data: { status: 'deleted' } }) };
  const service = new AccountService(http);
  assert.equal(await rx.firstValueFrom(service.delete('secret-pass')), 'deleted');
  assert.equal(request.method, 'DELETE');
  assert.equal(request.url, '/api/v1/account');
  assert.equal(JSON.stringify(request.options.body), JSON.stringify({ password: 'secret-pass' }));
  assert.equal(request.options.observe, 'response');
  http.request = reply(202, { data: { status: 'pending' } });
  assert.equal(await rx.firstValueFrom(service.delete('secret-pass')), 'pending');
  http.request = () => rx.throwError(() => new HttpErrorResponse({ status: 403 }));
  await assert.rejects(rx.firstValueFrom(service.delete('wrong')), (e) => e.status === 403);
});

test('local sign-out after deletion clears the session without calling the server', () => {
  const source = fs.readFileSync(
    path.join(__dirname, '../src/app/core/services/auth.service.ts'),
    'utf8',
  );
  const body = source.slice(source.indexOf('endSessionLocally'));
  const method = body.slice(0, body.indexOf('\n  }\n'));
  assert.match(method, /this\.revision\+\+/);
  assert.match(method, /this\.commit\(emptySession\(\)\)/);
  assert.doesNotMatch(method, /http|endServerSession|navigate/);
});

test('login page explains deleted and pending deletion without overstating', () => {
  const source = fs.readFileSync(
    path.join(__dirname, '../src/app/pages/login/login.component.ts'),
    'utf8',
  );
  assert.match(source, /deleted:\s*\n\s*'Аккаунт удалён\./);
  assert.match(source, /'deletion-pending':\s*\n\s*'Запрос на удаление принят, но очистка данных ещё не завершена/);
  assert.match(source, /Резервные копии и технические журналы могут храниться дольше/);
});

test('delete dialog needs explicit confirmation and a password, and handles both outcomes', () => {
  const source = fs.readFileSync(
    path.join(__dirname, '../src/app/shared/delete-account/delete-account.component.ts'),
    'utf8',
  );
  assert.match(source, /\[disabled\]="submitting \|\| !understood \|\| !password"/);
  assert.match(source, /if \(this\.submitting \|\| !this\.understood \|\| !this\.password\) return;/);
  assert.match(source, /outcome === 'deleted' \? 'deleted' : 'deletion-pending'/);
  assert.match(source, /\[locked\]="submitting"/);
  assert.match(source, /this\.password = '';\s*\n\s*this\.submitting = false;/);
});

test('delete entry point is on both profile pages', () => {
  for (const file of [
    'src/app/pages/team/academic-profile.component.html',
    'src/app/pages/profile/profile.component.html',
  ])
    assert.match(fs.readFileSync(path.join(__dirname, '..', file), 'utf8'), /<app-delete-account \/>/);
});
