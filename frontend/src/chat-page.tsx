import { createContext } from 'preact'
import { useState, useRef } from 'preact/hooks'
import { BotIcon, HistoryIcon, ImagePlusIcon, XIcon, SendIcon } from './components/icons'
import { ContextVal, initStore, StoreI, SetState, useStore } from './utils/global-state'
import useAICompletion from '#/utils/hooks.tsx'
import { Loader } from '#/components/loader.tsx'
import { ChatHistory } from '#/chat-history.tsx'
import { models, SelectModel } from '#/select-models.tsx'
import { Messages } from '#/messages.tsx'

export const Ctx = createContext<ContextVal<typeof chatStore>>(null)

export interface ChatStore {
  seletedChatId: number | undefined
  selectedModel: string
  selectedProvider: string
  setChat: (chatId: number | undefined) => void
  setModelAndProvider: (modelId: string, provider: string) => void
}

function chatStore(set: SetState<ChatStore>): StoreI<ChatStore> {
  return {
    seletedChatId: undefined,
    selectedModel: models[0].modelId,
    selectedProvider: models[0].provider,
    setModelAndProvider: (modelId, provider) => {
      set((prev) => ({ ...prev, selectedModel: modelId, selectedProvider: provider }))
    },
    setChat: (chatId) => set((prev) => ({ ...prev, seletedChatId: chatId })),
  }
}

function handleTextInputSize(e: Event) {
  const inputElement = e.target as HTMLTextAreaElement
  inputElement.style.height = 'auto'
  inputElement.style.height = inputElement.scrollHeight + 'px'
  e.stopPropagation()
}

export const ChatInterface = () => {
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [selectedImages, setSelectedImages] = useState<string[]>([])
  const storeValue = initStore<StoreI<ChatStore>>(chatStore)
  const selectedChatId = useStore(storeValue, (s) => s.seletedChatId)
  const { setChat } = storeValue.getState()
  const { mutate, isPending } = useAICompletion(selectedChatId, storeValue)

  function handleImageSelect(e: Event) {
    const files = (e.target as HTMLInputElement).files
    if (!files || files.length === 0) return
    for (const file of Array.from(files)) {
      if (!file.type.startsWith('image/')) {
        alert('Please select image files only')
        return
      }
      const reader = new FileReader()
      reader.onload = (e) => {
        const result = e.target?.result as string
        setSelectedImages((prev) => [...prev, result])
      }
      reader.readAsDataURL(file)
    }
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  function handleSend() {
    if (inputRef.current == null) return
    const input = inputRef.current!.value.trim()
    inputRef.current.value = ''
    if (!input && selectedImages.length === 0) return
    mutate({ message: input || 'What is in this image?', imageData: selectedImages.length > 0 ? selectedImages : undefined })
    setSelectedImages([])
  }

  function removeImage(index: number) {
    setSelectedImages((prev) => prev.filter((_, i) => i !== index))
  }

  return (
    <Ctx.Provider value={storeValue}>
      <div className="flex h-screen dark bg-gray-950">
        <div className="w-64 border-r p-4 bg-gray-900 border-gray-700 h-screen overflow-y-auto">
          <div className="flex items-center gap-2 mb-6">
            <BotIcon className="h-6 w-6 text-white" />
            <h1 className="text-xl font-bold text-white">AI Chat</h1>
          </div>

          <button
            type="button"
            className="w-full text-white focus:ring-4 font-medium rounded-lg text-sm px-5 py-2.5 me-2 mb-4 bg-blue-600 hover:bg-blue-700 focus:outline-hidden focus:ring-blue-800"
            onClick={() => setChat(undefined)}>
            New Chat
          </button>

          <div className="flex items-center gap-2 mb-4 text-white">
            <HistoryIcon className="h-4 w-4" />
            <span className="font-medium">Chat History</span>
          </div>

          <ChatHistory selectedChatId={selectedChatId} />
        </div>

        <div className="flex-1 flex flex-col h-screen max-w-[calc(100vw-16rem)]">
          <SelectModel chatId={selectedChatId} />

          <Messages chatId={selectedChatId} />

          {isPending && <Loader />}

          <div className="p-4 border-t bg-gray-900 border-gray-500">
            {selectedImages.length > 0 && (
              <div className="mb-2 flex gap-2 flex-wrap">
                {selectedImages.map((img, index) => (
                  <div key={index} className="relative inline-block">
                    <img src={img} alt="Selected" className="max-h-32 rounded-lg border border-gray-500" />
                    <button
                      onClick={() => removeImage(index)}
                      className="absolute -top-2 -right-2 bg-red-500 rounded-full p-1 hover:bg-red-600">
                      <XIcon className="h-4 w-4 text-white" />
                    </button>
                  </div>
                ))}
              </div>
            )}
            <div className="flex gap-2 items-center">
              <input type="file" ref={fileInputRef} onChange={handleImageSelect} accept="image/*" multiple className="hidden" />
              <button
                onClick={() => fileInputRef.current?.click()}
                className="border border-gray-500 rounded-sm p-4 hover:bg-gray-800"
                title="Upload image">
                <ImagePlusIcon className="h-4 w-4 text-white" />
              </button>
              <textarea
                className="bg-transparent text-sm text-white placeholder:text-zinc-400 focus-visible:ring-gray-300 rounded-md border p-2 focus-visible:ring-1 disabled:opacity-50 h-auto max-h-96 w-full overflow-auto resize-none flex-1"
                placeholder="Type your message..."
                onInput={handleTextInputSize}
                ref={inputRef}
              />
              <button className="border border-gray-500 rounded-sm px-6 size-16" onClick={handleSend}>
                <SendIcon className="h-4 w-4 text-white" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </Ctx.Provider>
  )
}
