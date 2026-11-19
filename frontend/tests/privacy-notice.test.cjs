const test = require('node:test'),
  assert = require('node:assert/strict');
const fs = require('node:fs'),
  path = require('node:path'),
  vm = require('node:vm'),
  ts = require('typescript');
const read = (...parts) => fs.readFileSync(path.join(__dirname, '..', '..', ...parts), 'utf8');
function notice() {
  const module = { exports: {} };
  vm.runInNewContext(
    ts.transpileModule(read('frontend/src/app/shared/privacy-notice.ts'), {
      compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
    }).outputText,
    { module, exports: module.exports },
  );
  return module.exports;
}

test('frontend notice version equals the version the server stores', () => {
  const go = read('auth/internal/app/domain/services/service.go');
  const server = /const ConsentNoticeVersion = "([^"]+)"/.exec(go)?.[1];
  assert.ok(server, 'ConsentNoticeVersion not found');
  assert.equal(notice().PRIVACY_NOTICE_VERSION, server);
});

test('notice covers data, purpose, visibility and deletion', () => {
  const { PRIVACY_NOTICE } = notice();
  assert.equal(
    JSON.stringify(PRIVACY_NOTICE.map((s) => s.title)),
    JSON.stringify([
      'Какие данные мы собираем',
      'Зачем это нужно',
      'Кто что видит',
      'Как удалить аккаунт',
    ]),
  );
  for (const section of PRIVACY_NOTICE) assert.ok(section.items.length > 0);
});

test('notice text matches docs/privacy-notice.md', () => {
  const doc = read('docs/privacy-notice.md');
  const { PRIVACY_NOTICE, PRIVACY_NOTICE_VERSION } = notice();
  assert.ok(doc.includes(PRIVACY_NOTICE_VERSION));
  for (const section of PRIVACY_NOTICE) {
    assert.ok(doc.includes(section.title), section.title);
    for (const item of section.items) assert.ok(doc.includes(item), item);
  }
});

test('registration page shows the notice and keeps consent unchecked by default', () => {
  const html = read('frontend/src/app/pages/register/register.component.html');
  const component = read('frontend/src/app/pages/register/register.component.ts');
  assert.match(html, /\*ngFor="let section of notice"/);
  assert.match(html, /\[disabled\]="loading \|\| !form\.valid \|\| !consentAccepted"/);
  assert.match(component, /consentAccepted = false;/);
  assert.match(component, /if \(!this\.consentAccepted\) \{/);
});

test('notice does not promise what the system does not do', () => {
  const text = JSON.stringify(notice().PRIVACY_NOTICE);
  assert.doesNotMatch(text, /немедленно|сразу удал|навсегда стерт|никому не передаём/i);
  assert.match(text, /Резервные копии и технические журналы могут храниться дольше/);
});
