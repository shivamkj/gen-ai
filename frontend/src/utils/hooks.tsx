import { useMutation, invalidateQuery } from '#/utils/query.ts'
import { ChatStore } from '../chat-page'
import { StoreI } from './global-state'
import { StoreApi, useStore } from './global-state'

export const baseUrl = `http://${window.location.host}`

export default function useAICompletion(chatId: number | undefined, chatStore: StoreApi<StoreI<ChatStore>>) {
  const model = useStore(chatStore, (s) => s.selectedModel)
  const provider = useStore(chatStore, (s) => s.selectedProvider)
  const { setChat } = chatStore.getState()

  const { mutate, error, isPending } = useMutation({
    mutationFn: async (data: { message: string; imageData?: string[] }) => {
      if (chatId == null) {
        const url = new URL('/api/chats/start', baseUrl)
        const body = JSON.stringify({ message: data.message, model, provider, imageData: data.imageData })
        const resp = await fetch(url, { method: 'POST', body, headers: { 'Content-Type': 'application/json' } })
        return resp.json()
      }
      const url = new URL(`/api/chats/${chatId}/reply`, baseUrl)
      const body = JSON.stringify({ message: data.message, imageData: data.imageData })
      return fetch(url, { method: 'POST', body, headers: { 'Content-Type': 'application/json' } }).then((resp) =>
        resp.json()
      )
    },
    onError: (error, data) => {
      console.error(error)
      console.log('==data==')
      console.log(data.message)
      console.log('==data==')
      alert(`Unexpected Error Occured: ${error?.name}: ${error?.message}`)
    },
    onSuccess: (data) => {
      const createdChatId = data?.chat?.id
      if (createdChatId != null) {
        setChat(createdChatId)
        invalidateQuery(['chatHistory'])
      }
      invalidateQuery(['chat', chatId])
    },
  })

  return { mutate, error, isPending } as const
}
