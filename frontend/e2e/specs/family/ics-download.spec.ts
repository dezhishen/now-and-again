/**
 * ICS 下载与内容验证测试
 *
 * 覆盖：
 * - API Key 认证的 ICS 下载
 * - Basic Auth 认证的 ICS 下载
 * - 无认证被拒绝
 * - 错误凭据被拒绝
 * - iCalendar 格式验证
 */
import { test, expect } from '@playwright/test';
import * as api from '../utils/api';

const API_BASE = 'http://localhost:8080';

let familyId = '';
let rawApiKey = '';
let apiKeyId = '';

test.beforeAll(async () => {
  await api.login('admin', '12345678');
  const res = await api.listMyFamilies();
  const families = Array.isArray((res.data as any)?.data)
    ? (res.data as any).data
    : res.data || [];
  familyId = families[0]?.id || '';
  if (!familyId) {
    const c = await api.createFamily('E2E-ICS-' + Date.now());
    familyId = c.familyId;
  }
  expect(familyId).toBeTruthy();

  // Create an API key for ICS feed auth
  const keyRes = await api.request<any>('POST', '/users/me/api-keys', {
    name: 'E2E-ICS-Key-' + Date.now(),
    scopes: ['read'],
  });
  rawApiKey = keyRes.data?.data?.api_key?.raw_key || keyRes.data?.api_key?.raw_key || '';
  apiKeyId = keyRes.data?.data?.api_key?.id || keyRes.data?.api_key?.id || '';
});

// ─── Helper: create ICS feed via API ─────────────────────────────
async function createIcsFeed(payload: Record<string, any>) {
  api.setFamilyId(familyId);
  const res = await api.request<any>('POST', '/ics-feeds', payload);
  return res;
}

// ─── Helper: fetch ICS content directly (bypasses JWT) ────────────
async function fetchIcs(
  accessToken: string,
  opts?: { key?: string; username?: string; password?: string },
): Promise<{ status: number; body: string; contentType: string | null }> {
  const url = `${API_BASE}/api/ics/${accessToken}${opts?.key ? `?key=${encodeURIComponent(opts.key)}` : ''}`;
  const headers: Record<string, string> = {};
  if (opts?.username && opts?.password) {
    const encoded = Buffer.from(`${opts.username}:${opts.password}`).toString('base64');
    headers['Authorization'] = `Basic ${encoded}`;
  }
  const res = await fetch(url, { headers });
  return {
    status: res.status,
    body: await res.text(),
    contentType: res.headers.get('content-type'),
  };
}

// ─── Tests ────────────────────────────────────────────────────────

test.describe('ICS 下载', () => {

  test('API Key: 创建 Feed → 下载 ICS → 验证内容', async () => {
    expect(rawApiKey).toBeTruthy();
    expect(apiKeyId).toBeTruthy();

    // Create ICS feed with API key auth
    const feedRes = await createIcsFeed({
      name: 'API Key Feed',
      description: 'Test ICS feed with API key auth',
      filter_days: 7,
      filter_type: 'all',
      auth_type: 'api_key',
      api_key_id: apiKeyId,
    });
    const feed = (feedRes.data as any)?.data || feedRes.data;
    expect(feed?.id).toBeTruthy();

    // Extract access token from ics_url: "/api/ics/<token>/<slug>.ics"
    const icsUrl: string = feed.ics_url || '';
    const tokenMatch = icsUrl.match(/\/api\/ics\/([^/]+)/);
    expect(tokenMatch).toBeTruthy();
    const accessToken = tokenMatch![1];

    // Download ICS
    const ics = await fetchIcs(accessToken, { key: rawApiKey });
    expect(ics.status).toBe(200);
    expect(ics.contentType).toContain('text/calendar');

    // Verify iCalendar structure
    expect(ics.body).toContain('BEGIN:VCALENDAR');
    expect(ics.body).toContain('VERSION:2.0');
    expect(ics.body).toContain('PRODID:-//NowAndAgain//Family Calendar//EN');
    expect(ics.body).toContain('END:VCALENDAR');
    expect(ics.body).toContain('CALSCALE:GREGORIAN');
    expect(ics.body).toContain('X-WR-CALNAME:API Key Feed');
  });

  test('Basic Auth: 创建 Feed → 下载 ICS → 验证内容', async () => {
    // Create ICS feed with Basic auth
    const feedRes = await createIcsFeed({
      name: 'Basic Auth Feed',
      description: 'Test ICS feed with Basic auth',
      filter_days: 7,
      filter_type: 'all',
      auth_type: 'basic',
      app_password: 'test1234',
    });
    const feed = (feedRes.data as any)?.data || feedRes.data;
    expect(feed?.id).toBeTruthy();

    const icsUrl: string = feed.ics_url || '';
    const tokenMatch = icsUrl.match(/\/api\/ics\/([^/]+)/);
    expect(tokenMatch).toBeTruthy();
    const accessToken = tokenMatch![1];

    // The app_username should be 'admin' (derived from the user account)
    const ics = await fetchIcs(accessToken, { username: 'admin', password: 'test1234' });
    expect(ics.status).toBe(200);
    expect(ics.contentType).toContain('text/calendar');

    // Verify iCalendar structure
    expect(ics.body).toContain('BEGIN:VCALENDAR');
    expect(ics.body).toContain('VERSION:2.0');
    expect(ics.body).toContain('END:VCALENDAR');
    expect(ics.body).toContain('X-WR-CALNAME:Basic Auth Feed');
  });

  test('无认证 → 返回 401', async () => {
    // Create a feed first
    const feedRes = await createIcsFeed({
      name: 'No Auth Feed',
      filter_days: 7,
      auth_type: 'api_key',
      api_key_id: apiKeyId,
    });
    const feed = (feedRes.data as any)?.data || feedRes.data;
    const icsUrl: string = feed.ics_url || '';
    const tokenMatch = icsUrl.match(/\/api\/ics\/([^/]+)/);
    expect(tokenMatch).toBeTruthy();

    // Download without any auth
    const ics = await fetchIcs(tokenMatch![1]);
    expect(ics.status).toBe(401);
  });

  test('错误 API Key → 返回 401', async () => {
    // Create a feed
    const feedRes = await createIcsFeed({
      name: 'Wrong Key Feed',
      filter_days: 7,
      auth_type: 'api_key',
      api_key_id: apiKeyId,
    });
    const feed = (feedRes.data as any)?.data || feedRes.data;
    const icsUrl: string = feed.ics_url || '';
    const tokenMatch = icsUrl.match(/\/api\/ics\/([^/]+)/);
    expect(tokenMatch).toBeTruthy();

    // Download with wrong key
    const ics = await fetchIcs(tokenMatch![1], { key: 'wrong-key-12345' });
    expect(ics.status).toBe(401);
  });

  test('错误 Basic Auth 密码 → 返回 401', async () => {
    // Create a Basic auth feed
    const feedRes = await createIcsFeed({
      name: 'Wrong Basic Feed',
      filter_days: 7,
      auth_type: 'basic',
      app_password: 'test1234',
    });
    const feed = (feedRes.data as any)?.data || feedRes.data;
    const icsUrl: string = feed.ics_url || '';
    const tokenMatch = icsUrl.match(/\/api\/ics\/([^/]+)/);
    expect(tokenMatch).toBeTruthy();

    // Download with wrong password
    const ics = await fetchIcs(tokenMatch![1], { username: 'admin', password: 'wrongpass' });
    expect(ics.status).toBe(401);
  });

  test('已创建任务后 ICS 包含 VEVENT', async () => {
    // Create a task first
    const taskName = 'E2E-ICS-Task-' + Date.now();
    await api.createTask(familyId, { name: taskName, kind: 'simple' });

    // Create ICS feed
    const feedRes = await createIcsFeed({
      name: 'Task ICS Feed',
      filter_days: 7,
      auth_type: 'api_key',
      api_key_id: apiKeyId,
    });
    const feed = (feedRes.data as any)?.data || feedRes.data;
    const icsUrl: string = feed.ics_url || '';
    const tokenMatch = icsUrl.match(/\/api\/ics\/([^/]+)/);
    expect(tokenMatch).toBeTruthy();

    // Download
    const ics = await fetchIcs(tokenMatch![1], { key: rawApiKey });
    expect(ics.status).toBe(200);

    // ICS should contain the task as a VEVENT
    expect(ics.body).toContain('BEGIN:VEVENT');
    expect(ics.body).toContain('END:VEVENT');
    expect(ics.body).toContain(`SUMMARY:${taskName}`);
  });
});
