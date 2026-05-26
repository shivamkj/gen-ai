import { useAction, useStoreX } from '#/utils/global-state.ts'
import { Ctx } from '#/chat-page.tsx'

const modelsByProvider: Record<string, { label: string; models: { name: string; modelId: string }[] }> = {
  bedrock: {
    label: 'AWS Bedrock',
    models: [
      { name: 'Claude 4.6 Sonnet', modelId: 'us.anthropic.claude-sonnet-4-6' },
      { name: 'Claude 4.5 Sonnet', modelId: 'us.anthropic.claude-sonnet-4-5-20250929-v1:0' },
    ],
  },
  deepseek: {
    label: 'Deepseek',
    models: [{ name: 'Deepseek v3', modelId: 'deepseek-chat' }],
  },
}

export const models = Object.entries(modelsByProvider).flatMap(([provider, { models: ms }]) =>
  ms.map((m) => ({ ...m, provider }))
)

const providers = Object.entries(modelsByProvider).map(([id, { label }]) => ({ id, label }))

const selectClass =
  'border text-sm rounded-lg block p-2.5 bg-gray-700 border-gray-600 placeholder-gray-400 text-white focus:ring-blue-500 focus:border-blue-500'

export function SelectModel({ chatId: selectedChatId }: { chatId: number | undefined }) {
  const selectedModel = useStoreX(Ctx, (s) => s.selectedModel)
  const selectedProvider = useStoreX(Ctx, (s) => s.selectedProvider)
  const { setModelAndProvider } = useAction(Ctx)

  const modelsForProvider = models.filter((m) => m.provider === selectedProvider)

  function handleProviderChange(e: Event) {
    const provider = (e.target as HTMLSelectElement).value
    const firstModel = models.find((m) => m.provider === provider)!
    setModelAndProvider(firstModel.modelId, provider)
  }

  function handleModelChange(e: Event) {
    const modelId = (e.target as HTMLSelectElement).value
    setModelAndProvider(modelId, selectedProvider)
  }

  return (
    <div className="p-4 border-b flex items-center gap-3 bg-gray-800 border-gray-700">
      {selectedChatId == undefined && (
        <>
          <select className={selectClass} value={selectedProvider} onChange={handleProviderChange}>
            {providers.map(({ id, label }) => (
              <option key={id} value={id}>
                {label}
              </option>
            ))}
          </select>

          <select className={selectClass} value={selectedModel} onChange={handleModelChange}>
            {modelsForProvider.map((m) => (
              <option key={m.modelId} value={m.modelId}>
                {m.name}
              </option>
            ))}
          </select>
        </>
      )}
    </div>
  )
}
