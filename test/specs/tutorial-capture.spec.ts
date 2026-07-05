import { test, expect, type Page } from '@playwright/test'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const SHOT_DIR = path.resolve(__dirname, '../../doc/tutorial/images')

function shotPath(tutorial: string, name: string): string {
  const dir = path.join(SHOT_DIR, tutorial)
  fs.mkdirSync(dir, { recursive: true })
  return path.join(dir, name)
}

function uniqueFamilyName(): string {
  return `阳光家庭-${Date.now()}`
}

async function loginAndEnterFamily(page: Page) {
  await page.goto('/login')
  await page.locator('[data-testid="login-username"]').fill('admin')
  await page.locator('[data-testid="login-password"]').fill('12345678')

  for (let i = 0; i < 5; i++) {
    await page.locator('[data-testid="login-submit"]').click()
    try {
      await expect(page).not.toHaveURL(/\/login/, { timeout: 4000 })
      break
    } catch {
      if (i === 4) throw new Error('login failed after retries')
      await page.waitForTimeout(1000)
    }
  }

  if (page.url().includes('/family')) return

  await expect(page.getByRole('heading', { name: /家庭|Home/ })).toBeVisible()
  await page.waitForTimeout(400)

  let enterButtons = page.locator('[data-testid="family-enter-btn"]')
  let familyCards = page.locator('[data-testid="family-card"]')

  if (await enterButtons.count() === 0 && await familyCards.count() === 0) {
    const createToggle = page.locator('[data-testid="family-create-toggle"]').first()
    const createToggleFallback = page.getByRole('button', { name: /创建家庭|创建/ }).first()
    if (await createToggle.count() > 0 || await createToggleFallback.count() > 0) {
      if (await createToggle.count() > 0) await createToggle.click()
      else await createToggleFallback.click()

      const familyName = uniqueFamilyName()
      const familyNameInput = page.locator('[data-testid="family-name-input"]').first()
      if (await familyNameInput.count() > 0) await familyNameInput.fill(familyName)
      else await page.getByPlaceholder(/家庭名称|Family name/).first().fill(familyName)

      const familyCreateSubmit = page.locator('[data-testid="family-create-submit"]').first()
      if (await familyCreateSubmit.count() > 0) await familyCreateSubmit.click()
      else await page.getByRole('button', { name: /创建|Create/ }).first().click()

      await page.waitForTimeout(800)
      enterButtons = page.locator('[data-testid="family-enter-btn"]')
      familyCards = page.locator('[data-testid="family-card"]')
    }
  }

  if (await enterButtons.count() > 0) {
    for (let i = 0; i < 3 && !page.url().includes('/family'); i++) {
      try { await page.locator('[data-testid="family-enter-btn"]').first().click({ timeout: 5000 }) } catch { await page.waitForTimeout(300) }
    }
  }

  if (!page.url().includes('/family') && await familyCards.count() > 0) {
    for (let i = 0; i < 3 && !page.url().includes('/family'); i++) {
      try { await familyCards.first().click({ timeout: 5000 }) } catch { await page.waitForTimeout(300) }
    }
  }

  if (!page.url().includes('/family') && await page.locator('[data-testid="family-enter-btn"]').count() > 0) {
    await page.locator('[data-testid="family-enter-btn"]').first().click()
  }

  await expect(page).toHaveURL(/\/family/)
  await page.waitForTimeout(300)
}

// ──────────────────────────────────────────────────────────────────
// 1. 家庭管理教程
// ──────────────────────────────────────────────────────────────────
test('教程截图：家庭管理——创建家庭与邀请成员', async ({ page }) => {
  // Login fresh (no family yet)
  await page.goto('/login')
  await page.locator('[data-testid="login-username"]').fill('admin')
  await page.locator('[data-testid="login-password"]').fill('12345678')
  await page.locator('[data-testid="login-submit"]').click()
  await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 })

  await page.waitForTimeout(300)

  // If already in a family, switch to fresh account or navigate to home
  if (!page.url().includes('/families')) {
    await page.goto('/')
    await page.waitForTimeout(500)
  }

  // Screenshot: family list / create page
  await page.screenshot({ path: shotPath('family', '01-family-list-create.png'), fullPage: true })

  // Create family
  const createBtn = page.getByRole('button', { name: /创建家庭|创建/ }).first()
  if (await createBtn.count() > 0) {
    await createBtn.click()
    const nameInput = page.locator('[data-testid="family-name-input"]').first()
    if (await nameInput.count() > 0) await nameInput.fill(uniqueFamilyName())
    else await page.getByPlaceholder(/家庭名称|Family name/).first().fill(uniqueFamilyName())

    const submit = page.locator('[data-testid="family-create-submit"]').first()
    if (await submit.count() > 0) await submit.click()
    else await page.getByRole('button', { name: /创建|Create/ }).first().click()
    await page.waitForTimeout(600)
  }

  // Enter family
  const enterBtn = page.locator('[data-testid="family-enter-btn"]').first()
  if (await enterBtn.count() > 0) {
    await enterBtn.click()
    await expect(page).toHaveURL(/\/family/, { timeout: 5000 })
  }

  // Go to dashboard overview tab
  const overviewBtn = page.getByText(/概况|Overview/).first()
  if (await overviewBtn.count() > 0) await overviewBtn.click()
  await page.waitForTimeout(300)

  // Screenshot: dashboard overview with invite code
  await page.screenshot({ path: shotPath('family', '02-dashboard-overview.png'), fullPage: true })

  // Go to members tab
  await page.locator('[data-testid="family-nav-members"]').first().click()
  await expect(page.getByText(/成员|Members/).first()).toBeVisible()
  await page.waitForTimeout(300)

  // Screenshot: members list
  await page.screenshot({ path: shotPath('family', '03-members-list.png'), fullPage: true })
})

// ──────────────────────────────────────────────────────────────────
// 2. 巡检任务教程
// ──────────────────────────────────────────────────────────────────
test('教程截图：巡检任务——创建触发与完成', async ({ page }) => {
  await loginAndEnterFamily(page)

  // Navigate to tasks
  await page.locator('[data-testid="family-nav-tasks"]').first().click()
  await expect(page.locator('[data-testid="task-create-btn"]')).toBeVisible()

  // Create inspection task
  await page.locator('[data-testid="task-create-btn"]').click()

  await page.locator('[data-testid="task-kind"]').selectOption('inspection')
  await page.locator('[data-testid="task-name"]').fill('房间巡检')

  // Add check item
  await page.locator('[data-testid="check-item-add"]').click()
  await page.locator('[data-testid="check-item-name-input"]').first().fill('门窗状态')

  // Add second check item
  await page.locator('[data-testid="check-item-add"]').click()
  await page.locator('[data-testid="check-item-name-input"]').nth(1).fill('地面卫生')

  // Screenshot: inspection task creation form
  await page.screenshot({ path: shotPath('inspection', '01-inspection-create-form.png'), fullPage: true })

  await page.locator('[data-testid="task-submit"]').click()
  const card = page.locator('[data-testid="task-card"][data-task-name="房间巡检"]').first()
  await expect(card).toBeVisible()

  // Screenshot: inspection task card
  await page.screenshot({ path: shotPath('inspection', '02-inspection-task-card.png'), fullPage: true })

  // Trigger
  await card.locator('[data-testid="task-trigger-btn"]').click()
  await page.locator('[data-testid="family-nav-dashboard"]').first().click()

  const todoCard = page.locator('[data-testid="todo-card"][data-task-name="房间巡检"]').first()
  await expect(todoCard).toBeVisible()

  // Screenshot: dashboard with inspection todo
  await page.screenshot({ path: shotPath('inspection', '03-inspection-todo.png'), fullPage: true })

  // Open inspection detail
  await todoCard.locator('[data-testid="inspect-open-btn"]').click()
  await expect(page.getByRole('button', { name: '正常' }).first()).toBeVisible()

  await page.waitForTimeout(200)

  // Screenshot: inspection detail with check items
  await page.screenshot({ path: shotPath('inspection', '04-inspection-detail.png'), fullPage: true })

  // Complete the inspection: mark both items as normal
  await page.getByRole('button', { name: '正常' }).first().click()
  await page.getByRole('button', { name: '正常' }).nth(1).click()
  await page.locator('[data-testid="inspect-submit-btn"]').click()
  await expect(todoCard).not.toBeVisible({ timeout: 10000 })

  // Screenshot: dashboard after inspection completed
  await page.screenshot({ path: shotPath('inspection', '05-inspection-done.png'), fullPage: true })
})

// ──────────────────────────────────────────────────────────────────
// 3. 日历教程
// ──────────────────────────────────────────────────────────────────
test('教程截图：日历与大屏', async ({ page }) => {
  await loginAndEnterFamily(page)

  // Navigate to ICS/Calendar tab
  await page.locator('[data-testid="family-nav-ics"]').first().click()
  await page.waitForTimeout(500)

  // Screenshot: ICS feeds config
  await page.screenshot({ path: shotPath('calendar', '01-ics-feeds.png'), fullPage: true })

  // Open embed calendar
  await page.locator('[data-testid="family-nav-dashboard"]').first().click()

  // Open calendar as new tab for full-screen view
  await page.evaluate(() => window.open('/calendar', '_blank'))
  await page.waitForTimeout(500)

  // Get all pages and switch to the calendar one
  const context = page.context()
  const calendarPage = context.pages().length > 1 ? context.pages()[1] : page

  if (context.pages().length > 1) {
    await calendarPage.bringToFront()
    await calendarPage.waitForTimeout(800)
    await calendarPage.screenshot({ path: shotPath('calendar', '02-calendar-fullscreen.png'), fullPage: true })
    await calendarPage.close()
  }
})
