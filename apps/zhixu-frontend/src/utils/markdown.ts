import MarkdownIt from 'markdown-it'

const markdown = new MarkdownIt({ html: false, linkify: false, typographer: false })
const defaultLink = markdown.renderer.rules.link_open
markdown.renderer.rules.link_open = (tokens, index, options, env, self) => {
  const token = tokens[index]!
  const href = String(token.attrGet('href') || '')
  if (!/^https?:\/\//i.test(href) && !/^mailto:/i.test(href) && !href.startsWith('#'))
    token.attrSet('href', '#')
  token.attrSet('rel', 'noopener noreferrer')
  return defaultLink
    ? defaultLink(tokens, index, options, env, self)
    : self.renderToken(tokens, index, options)
}
// Imported documents must not load remote tracking images or authenticated local URLs.
markdown.renderer.rules.image = (tokens, index) =>
  `<span class="markdown-image-note">[图片：${markdown.utils.escapeHtml(tokens[index]?.content || '未加载')}，暂不加载]</span>`

export const renderMarkdown = (content: string) => markdown.render(content)
