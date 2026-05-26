import { useQuery, useMutation, invalidateQuery } from '#/utils/query.ts'
import { useEffect, useRef } from 'preact/hooks'
import { TrashIcon } from '#/components/icons.tsx'
import { ScrollArea } from './components/scroll-area'
import { Markdown } from './components/markdown'
import { Avatar } from './components/avatar'
import { baseUrl } from '#/utils/hooks.tsx'

export function Messages({ chatId }: { chatId: number | undefined }) {
  const { data } = useQuery({
    queryKey: ['chat', chatId] as const,
    enabled: chatId != null,
    queryFn: () => fetch(`${baseUrl}/api/chats/${chatId}/messages`).then((res) => res.json()),
  })

  const { mutate: deleteMessage } = useMutation({
    mutationFn: async (id: number) => {
      const response = await fetch(`${baseUrl}/api/messages/${id}`, { method: 'DELETE' })
      return response.json()
    },
    onSuccess: () => {
      invalidateQuery(['chat', chatId])
    },
  })

  const scrollAreaRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!scrollAreaRef.current) return
    const scrollContainer = scrollAreaRef.current.querySelector('[data-scrollable]')
    if (scrollContainer) scrollContainer.scrollTop = scrollContainer.scrollHeight
  }, [data])

  function deleteChat(messageId: number) {
    const confirmed = window.confirm('Are you sure?')
    if (confirmed) deleteMessage(messageId)
  }

  if (data == null) return <div className="flex-1 text-2xl"></div>

  return (
    <ScrollArea outerClass="flex-1" innerClass="p-4" outerRef={scrollAreaRef}>
      {data.map((message: any, index: any) => (
        <div
          key={index}
          className={`relative mb-4 flex ${message.role == 'assistant' ? 'flex-row-reverse justify-end' : ''}`}>
          <div
            className={`border-zinc-800 bg-gray-900 shadow-lg text-zinc-50 rounded-xl border p-6 overflow-x-hidden${message.role == 'user' ? ' bg-blue-500 text-white ml-auto' : ''}`}>
            {message.image_data && message.image_data.length > 0 && (
              <div className="mb-4 flex gap-2 flex-wrap">
                {message.image_data.map((img: string, i: number) => (
                  <img
                    key={i}
                    src={img}
                    alt="Uploaded"
                    className="max-w-full max-h-96 rounded-lg border border-gray-500"
                  />
                ))}
              </div>
            )}
            <Markdown content={message.content} />
            <hr />
            <div className="flex justify-between font-light">
              {message.output_token != null ? (
                <div className="font-light mr-10">
                  Tokens: {message.input_token} (Input), {message.output_token} (Output)
                </div>
              ) : (
                <div />
              )}
              <div>
                <TrashIcon className="text-red-700" onClick={() => deleteChat(message.id)} />
              </div>
            </div>
          </div>
          <Avatar className="mx-6" initials={message.role == 'user' ? 'YOU' : 'AI'} />
        </div>
      ))}
    </ScrollArea>
  )
}
