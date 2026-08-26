import { BLOCKS } from '../engine/blocks.ts'
import type { BlockId, CodeKind, CodeNode, ScriptStack, SoundName } from '../engine/types.ts'
import { makeNode, newId } from '../engine/world.ts'

interface ToolboxItem {
  kind: CodeKind
  label: string
  cat: 'event' | 'looks' | 'sound' | 'motion' | 'control' | 'sensing'
  hat?: boolean
}

const TOOLBOX: ToolboxItem[] = [
  { kind: 'hatStart', label: 'gdy gra się zaczyna', cat: 'event', hat: true },
  { kind: 'hatCoin', label: 'gdy zbierze monetę', cat: 'event', hat: true },
  { kind: 'hatTouch', label: 'gdy dotknie klocka', cat: 'event', hat: true },
  { kind: 'hatGoal', label: 'gdy dotrze do mety', cat: 'event', hat: true },
  { kind: 'hatDeath', label: 'gdy straci życie', cat: 'event', hat: true },
  { kind: 'hatAlways', label: 'zawsze', cat: 'event', hat: true },
  { kind: 'addScore', label: 'dodaj punkty', cat: 'sensing' },
  { kind: 'setMessage', label: 'ustaw wiadomość', cat: 'looks' },
  { kind: 'sound', label: 'zagraj dźwięk', cat: 'sound' },
  { kind: 'teleport', label: 'teleportuj na start', cat: 'motion' },
  { kind: 'win', label: 'wygraj', cat: 'control' },
  { kind: 'lose', label: 'przegrana', cat: 'control' },
  { kind: 'wait', label: 'poczekaj sekundy', cat: 'control' },
  { kind: 'ifScore', label: 'jeśli punkty ≥', cat: 'control' },
  { kind: 'repeat', label: 'powtórz razy', cat: 'control' },
]

interface BlockEditorProps {
  scripts: ScriptStack[]
  onChange: (scripts: ScriptStack[]) => void
}

export function BlockEditor({ scripts, onChange }: BlockEditorProps) {
  const addFromToolbox = (item: ToolboxItem) => {
    if (item.hat) {
      const stack: ScriptStack = {
        id: newId(),
        blocks: [nodeFromItem(item)],
      }
      onChange([...scripts, stack])
      return
    }
    if (scripts.length === 0) return
    const next = scripts.map((s, i) => {
      if (i !== scripts.length - 1) return s
      return insertNode(s, nodeFromItem(item))
    })
    onChange(next)
  }

  const updateStack = (id: string, blocks: CodeNode[]) => {
    onChange(scripts.map((s) => (s.id === id ? { ...s, blocks } : s)))
  }

  const removeStack = (id: string) => {
    onChange(scripts.filter((s) => s.id !== id))
  }

  return (
    <section className="code-panel" aria-label="Edytor klocków">
      <div className="toolbox">
        <h2>Klocki kodu</h2>
        {TOOLBOX.map((item) => (
          <button
            key={item.kind}
            type="button"
            className={`blk mini ${item.cat} ${item.hat ? 'hat' : ''}`}
            onClick={() => addFromToolbox(item)}
            data-testid={`code-${item.kind}`}
          >
            {item.label}
          </button>
        ))}
        <p className="hint">
          Najpierw żółte zdarzenie, potem pomarańczowe akcje. Klik dodaje klocek.
        </p>
      </div>
      <div className="workspace">
        {scripts.length === 0 ? (
          <p className="empty-code">
            Kliknij żółty klocek po lewej, żeby zacząć skrypt — tak jak w Scratchu.
          </p>
        ) : (
          scripts.map((stack) => (
            <StackView
              key={stack.id}
              stack={stack}
              onChange={(blocks) => updateStack(stack.id, blocks)}
              onRemove={() => removeStack(stack.id)}
            />
          ))
        )}
      </div>
    </section>
  )
}

function nodeFromItem(item: ToolboxItem): CodeNode {
  switch (item.kind) {
    case 'hatTouch':
      return makeNode(item.kind, { block: 8 })
    case 'addScore':
      return makeNode(item.kind, { amount: 10 })
    case 'setMessage':
      return makeNode(item.kind, { text: 'Cześć!' })
    case 'sound':
      return makeNode(item.kind, { sound: 'coin' })
    case 'wait':
      return makeNode(item.kind, { seconds: 1 })
    case 'ifScore':
      return makeNode(item.kind, { gte: 30, body: [] })
    case 'repeat':
      return makeNode(item.kind, { times: 3, body: [] })
    default:
      return makeNode(item.kind)
  }
}

function insertNode(stack: ScriptStack, node: CodeNode): ScriptStack {
  const last = stack.blocks[stack.blocks.length - 1]
  if (last && (last.kind === 'ifScore' || last.kind === 'repeat')) {
    return {
      ...stack,
      blocks: stack.blocks.map((b, i) =>
        i === stack.blocks.length - 1 ? { ...b, body: [...(b.body ?? []), node] } : b,
      ),
    }
  }
  return { ...stack, blocks: [...stack.blocks, node] }
}

function StackView({
  stack,
  onChange,
  onRemove,
}: {
  stack: ScriptStack
  onChange: (blocks: CodeNode[]) => void
  onRemove: () => void
}) {
  const patch = (index: number, next: CodeNode) => {
    onChange(stack.blocks.map((b, i) => (i === index ? next : b)))
  }
  const remove = (index: number) => {
    onChange(stack.blocks.filter((_, i) => i !== index))
  }

  return (
    <div className="stack">
      <button type="button" className="stack-del" onClick={onRemove} aria-label="Usuń skrypt">
        ×
      </button>
      {stack.blocks.map((node, i) => (
        <NodeView
          key={node.id}
          node={node}
          onChange={(n) => patch(i, n)}
          onRemove={() => remove(i)}
        />
      ))}
    </div>
  )
}

function NodeView({
  node,
  onChange,
  onRemove,
}: {
  node: CodeNode
  onChange: (node: CodeNode) => void
  onRemove: () => void
}) {
  const cat = category(node.kind)
  const hat = node.kind.startsWith('hat')
  const shaped = node.kind === 'ifScore' || node.kind === 'repeat'

  return (
    <div className={`blk ${cat} ${hat ? 'hat' : ''} ${shaped ? 'c-block' : ''}`}>
      <div className="blk-row">
        <NodeLabel node={node} onChange={onChange} />
        <button type="button" className="tiny" onClick={onRemove} aria-label="Usuń klocek">
          ×
        </button>
      </div>
      {shaped ? (
        <div className="c-body">
          {(node.body ?? []).map((child, i) => (
            <NodeView
              key={child.id}
              node={child}
              onChange={(n) => {
                const body = [...(node.body ?? [])]
                body[i] = n
                onChange({ ...node, body })
              }}
              onRemove={() => {
                const body = (node.body ?? []).filter((_, j) => j !== i)
                onChange({ ...node, body })
              }}
            />
          ))}
          <button
            type="button"
            className="add-inside"
            onClick={() =>
              onChange({
                ...node,
                body: [...(node.body ?? []), makeNode('addScore', { amount: 1 })],
              })
            }
          >
            + w środku
          </button>
        </div>
      ) : null}
    </div>
  )
}

function NodeLabel({
  node,
  onChange,
}: {
  node: CodeNode
  onChange: (node: CodeNode) => void
}) {
  switch (node.kind) {
    case 'hatStart':
      return <span>gdy gra się zaczyna</span>
    case 'hatCoin':
      return <span>gdy zbierze monetę</span>
    case 'hatGoal':
      return <span>gdy dotrze do mety</span>
    case 'hatDeath':
      return <span>gdy straci życie</span>
    case 'hatAlways':
      return <span>zawsze</span>
    case 'hatTouch':
      return (
        <span>
          gdy dotknie{' '}
          <select
            value={node.block ?? 8}
            onChange={(e) => onChange({ ...node, block: Number(e.target.value) as BlockId })}
          >
            {BLOCKS.filter((b) => b.id !== 0).map((b) => (
              <option key={b.id} value={b.id}>
                {b.name}
              </option>
            ))}
          </select>
        </span>
      )
    case 'addScore':
      return (
        <span>
          dodaj{' '}
          <input
            type="number"
            min={1}
            max={100}
            value={node.amount ?? 10}
            onChange={(e) => onChange({ ...node, amount: Number(e.target.value) })}
          />{' '}
          punktów
        </span>
      )
    case 'setMessage':
      return (
        <span>
          ustaw wiadomość{' '}
          <input
            type="text"
            maxLength={48}
            value={node.text ?? ''}
            onChange={(e) => onChange({ ...node, text: e.target.value })}
          />
        </span>
      )
    case 'sound':
      return (
        <span>
          zagraj{' '}
          <select
            value={node.sound ?? 'coin'}
            onChange={(e) => onChange({ ...node, sound: e.target.value as SoundName })}
          >
            <option value="coin">moneta</option>
            <option value="jump">skok</option>
            <option value="win">fanfara</option>
            <option value="lose">buuu</option>
            <option value="bump">buch</option>
            <option value="pop">pop</option>
          </select>
        </span>
      )
    case 'teleport':
      return <span>teleportuj na start</span>
    case 'win':
      return <span>wygraj</span>
    case 'lose':
      return <span>przegrana</span>
    case 'wait':
      return (
        <span>
          poczekaj{' '}
          <input
            type="number"
            min={0.1}
            max={5}
            step={0.1}
            value={node.seconds ?? 1}
            onChange={(e) => onChange({ ...node, seconds: Number(e.target.value) })}
          />{' '}
          s
        </span>
      )
    case 'ifScore':
      return (
        <span>
          jeśli punkty ≥{' '}
          <input
            type="number"
            min={0}
            max={999}
            value={node.gte ?? 0}
            onChange={(e) => onChange({ ...node, gte: Number(e.target.value) })}
          />
        </span>
      )
    case 'repeat':
      return (
        <span>
          powtórz{' '}
          <input
            type="number"
            min={1}
            max={20}
            value={node.times ?? 3}
            onChange={(e) => onChange({ ...node, times: Number(e.target.value) })}
          />{' '}
          razy
        </span>
      )
    default:
      return <span>{node.kind}</span>
  }
}

function category(kind: CodeKind): ToolboxItem['cat'] {
  if (kind.startsWith('hat')) return 'event'
  if (kind === 'setMessage') return 'looks'
  if (kind === 'sound') return 'sound'
  if (kind === 'teleport') return 'motion'
  if (kind === 'addScore') return 'sensing'
  return 'control'
}
