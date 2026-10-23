import { useCallback, useEffect, useRef, useState } from 'react'
import { AppHeader } from '@/components/app/AppHeader.tsx'
import { AnalysisProgress } from '@/features/evaluate/AnalysisProgress.tsx'
import { ImagePickerScreen } from '@/features/evaluate/ImagePickerScreen.tsx'
import { ResultScreen } from '@/features/evaluate/ResultScreen.tsx'
import { demoPhoto } from '@/data/landscapeSample.ts'
import type { SelectedPhoto } from '@/types/photo.ts'
import styles from './App.module.css'

type Phase = 'pick' | 'analyzing' | 'result'

export function App() {
  const [phase, setPhase] = useState<Phase>('pick')
  const [photo, setPhoto] = useState<SelectedPhoto | null>(null)
  const revokeObjectUrl = useRef<(() => void) | null>(null)

  useEffect(
    () => () => {
      revokeObjectUrl.current?.()
    },
    [],
  )

  const releasePhoto = () => {
    revokeObjectUrl.current?.()
    revokeObjectUrl.current = null
    setPhoto(null)
  }

  const handlePick = (file: File) => {
    revokeObjectUrl.current?.()
    const url = URL.createObjectURL(file)
    revokeObjectUrl.current = () => URL.revokeObjectURL(url)
    setPhoto({ name: file.name, url, alt: `Ảnh vừa chọn: ${file.name}`, isDemo: false })
  }

  const handleUseSample = () => {
    revokeObjectUrl.current?.()
    revokeObjectUrl.current = null
    setPhoto({ name: demoPhoto.name, url: demoPhoto.url, alt: demoPhoto.alt, isDemo: true })
  }

  const handleAnalyze = () => setPhase('analyzing')

  const handleAnalysisComplete = useCallback(() => setPhase('result'), [])

  const handleBackToPicker = () => {
    releasePhoto()
    setPhase('pick')
  }

  return (
    <div className={styles.shell}>
      <AppHeader />
      <main id="main" className={styles.main}>
        {phase === 'pick' && (
          <ImagePickerScreen
            photo={photo}
            onPick={handlePick}
            onClear={releasePhoto}
            onUseSample={handleUseSample}
            onAnalyze={handleAnalyze}
          />
        )}
        {phase === 'analyzing' && photo && (
          <AnalysisProgress photo={photo} onCancel={handleBackToPicker} onComplete={handleAnalysisComplete} />
        )}
        {phase === 'result' && photo && <ResultScreen photo={photo} onChangePhoto={handleBackToPicker} />}
      </main>
    </div>
  )
}
