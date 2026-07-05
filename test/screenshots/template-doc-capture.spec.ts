import { test, expect, type Page } from '@playwright/test'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const SHOT_DIR = path.resolve(__dirname, '../../doc/tutorial/images/template-quickstart')

function shotPath(name: string): string {
  return path.join(SHOT_DIR, name)
}

function uniqueFamilyName(): string {
  return `晨光之家-${Date.now()}`
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
      try {
        await page.locator('[data-testid="family-enter-btn"]').first().click({ timeout: 5000 })
      } catch {
        await page.waitForTimeout(300)
      }
    }
  }

  if (!page.url().includes('/family') && await familyCards.count() > 0) {
    for (let i = 0; i < 3 && !page.url().includes('/family'); i++) {
      try {
        await familyCards.first().click({ timeout: 5000 })
      } catch {
        await page.waitForTimeout(300)
      }
    }
  }

  if (!page.url().includes('/family') && await page.locator('[data-testid="family-enter-btn"]').count() > 0) {
    await page.locator('[data-testid="family-enter-btn"]').first().click()
  }

  await expect(page).toHaveURL(/\/family/)
}

test('文档截图：家庭内从模板创建任务并完成待办', async ({ page }) => {
  fs.mkdirSync(SHOT_DIR, { recursive: true })

  await loginAndEnterFamily(page)

  await page.locator('[data-testid="family-nav-tasks"]').first().click()
  await expect(page.locator('[data-testid="task-create-btn"]')).toBeVisible()
  await page.screenshot({ path: shotPath('01-family-task-entry.png'), fullPage: true })

  await page.locator('[data-testid="task-template-btn"]').click()
  await expect(page.getByText('选择模板')).toBeVisible()
  await page.screenshot({ path: shotPath('02-template-picker.png'), fullPage: true })

  await page.getByText('每日打卡', { exact: true }).first().click()
  await expect(page.getByText('填写参数')).toBeVisible()

  await page.getByPlaceholder('例如：晨会').fill('晨会打卡')
  await page.getByRole('button', { name: '预览' }).click()
  await expect(page.getByText('确认创建')).toBeVisible()
  await page.screenshot({ path: shotPath('03-template-preview.png'), fullPage: true })

  await page.getByRole('button', { name: '填充到任务表单' }).click()
  await expect(page.locator('[data-testid="task-name"]')).toHaveValue('每日晨会打卡')
  await page.screenshot({ path: shotPath('04-prefilled-task-form.png'), fullPage: true })

  await page.locator('[data-testid="task-submit"]').click()
  const taskCard = page.locator('[data-testid="task-card"][data-task-name="每日晨会打卡"]').first()
  await expect(taskCard).toBeVisible()

  await taskCard.locator('[data-testid="task-trigger-btn"]').click()
  await page.locator('[data-testid="family-nav-dashboard"]').first().click()

  const todoCard = page.locator('[data-testid="todo-card"][data-task-name="每日晨会打卡"]').first()
  await expect(todoCard).toBeVisible()
  await page.screenshot({ path: shotPath('05-dashboard-pending-todo.png'), fullPage: true })

  await todoCard.locator('[data-testid="todo-quick-done"]').click()
  await expect(todoCard).not.toBeVisible({ timeout: 10000 })
  await page.screenshot({ path: shotPath('06-dashboard-after-done.png'), fullPage: true })
})

test('文档截图：任务链创建与步骤推进', async ({ page }) => {
  fs.mkdirSync(SHOT_DIR, { recursive: true })

  await loginAndEnterFamily(page)

  // Navigate to tasks
  await page.locator('[data-testid="family-nav-tasks"]').first().click()
  await expect(page.locator('[data-testid="task-create-btn"]')).toBeVisible()

  // Create a chain task manually
  await page.locator('[data-testid="task-create-btn"]').click()
  await page.locator('[data-testid="task-kind"]').selectOption('chain')
  await page.locator('[data-testid="task-name"]').fill('新家入住巡检')

  // Add step 1: inspection
  await page.locator('[data-testid="chain-add-step"]').click()
  await page.locator('[data-testid="subtask-name-input"]').first().fill('水电检查')
  await page.locator('[data-testid="subtask-config-btn"]').first().click()
  await page.locator('[data-testid="subtask-kind-select"]').first().selectOption('inspection')
  await page.locator('[data-testid="check-item-add"]').click()
  await page.locator('[data-testid="check-item-name-input"]').first().fill('总水阀')
  await page.locator('[data-testid="subtask-confirm"]').click()

  // Add step 2: simple
  await page.locator('[data-testid="chain-add-step"]').click()
  await page.locator('[data-testid="subtask-name-input"]').nth(1).fill('门窗检查')

  // Screenshot: chain creation form with steps
  await page.screenshot({ path: shotPath('07-chain-create-form.png'), fullPage: true })

  await page.locator('[data-testid="task-submit"]').click()
  const card = page.locator('[data-testid="task-card"][data-task-name="新家入住巡检"]').first()
  await expect(card).toBeVisible()

  // Screenshot: chain task card with orange badge and steps summary
  await page.screenshot({ path: shotPath('08-chain-task-card.png'), fullPage: true })

  // Trigger chain todo
  await card.locator('[data-testid="task-trigger-btn"]').click()

  // Go to dashboard
  await page.locator('[data-testid="family-nav-dashboard"]').first().click()

  // Wait for chain todo to appear
  const chainTodo = page.locator('[data-testid="todo-card"][data-task-name="新家入住巡检"]').first()
  await expect(chainTodo).toBeVisible()
  await page.screenshot({ path: shotPath('09-chain-todo.png'), fullPage: true })

  // Complete root chain todo → advances to step 1
  await chainTodo.locator('[data-testid="chain-done-btn"]').click()
  await expect(chainTodo).not.toBeVisible({ timeout: 10000 })

  // Step 1 should now appear
  const step1Todo = page.locator('[data-testid="todo-card"][data-task-name="水电检查"]').first()
  await expect(step1Todo).toBeVisible()
  await page.screenshot({ path: shotPath('10-chain-step-progress.png'), fullPage: true })
})
