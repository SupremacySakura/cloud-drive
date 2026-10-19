import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import FilterMenu from '../../components/ui/FilterMenu.vue'

vi.mock('@iconify/vue', () => ({ Icon: { template: '<span />' } }))

describe('shared filter menu', () => {
  it('supports keyboard selection and Escape without a native select', async () => {
    const wrapper = mount(FilterMenu, {
      props: {
        modelValue: '',
        label: '筛选状态',
        options: [
          { key: '', label: '全部' },
          { key: 'ready', label: '已加入' },
        ],
      },
      attachTo: document.body,
    })
    await wrapper.get('[aria-haspopup="menu"]').trigger('click')
    await flushPromises()
    expect(document.activeElement?.textContent).toContain('全部')
    await wrapper.get('[role="menuitemradio"]').trigger('keydown', { key: 'ArrowDown' })
    expect(document.activeElement?.textContent).toContain('已加入')
    await wrapper.findAll('[role="menuitemradio"]')[1]!.trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['ready'])
    await wrapper.get('[aria-haspopup="menu"]').trigger('click')
    await flushPromises()
    await wrapper.get('[role="menuitemradio"]').trigger('keydown', { key: 'Escape' })
    expect(wrapper.get('[aria-haspopup="menu"]').attributes('aria-expanded')).toBe('false')
    expect(wrapper.find('select').exists()).toBe(false)
    wrapper.unmount()
  })
})
