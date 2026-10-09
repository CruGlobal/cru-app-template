import assert from "node:assert/strict";
import { createServer } from "node:http";
import type { AddressInfo } from "node:net";
import { after, before, beforeEach, test } from "node:test";
import { handler } from "./app.ts";

const server = createServer((req, res) => void handler(req, res));
let base = "";
before(async () => {
  await new Promise<void>((resolve) => server.listen(0, resolve));
  base = `http://localhost:${(server.address() as AddressInfo).port}`;
});
after(() => server.close());
beforeEach(() => {
  delete process.env.IAP_AUDIENCE;
  delete process.env.CRU_IAP_DEV_BYPASS_EMAIL;
});

test("/up is always open", async () => {
  assert.equal((await fetch(`${base}/up`)).status, 200);
});

test("without an assertion or the dev bypass, everything else is a 401", async () => {
  assert.equal((await fetch(`${base}/`)).status, 401);
  assert.equal((await fetch(`${base}/nope`)).status, 401);
});

test("without IAP_AUDIENCE the dev bypass names the user", async () => {
  process.env.CRU_IAP_DEV_BYPASS_EMAIL = "dev@example.com";
  const res = await fetch(`${base}/`);
  assert.equal(res.status, 200);
  assert.equal(await res.text(), "Hello, dev@example.com 👋");
});

test("with IAP_AUDIENCE the dev bypass is ignored", async () => {
  process.env.IAP_AUDIENCE = "/projects/1/global/backendServices/2";
  process.env.CRU_IAP_DEV_BYPASS_EMAIL = "dev@example.com";
  assert.equal((await fetch(`${base}/`)).status, 401);
  assert.equal((await fetch(`${base}/up`)).status, 200);
});
