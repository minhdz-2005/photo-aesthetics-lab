import { useState } from 'react'
import { BeforeAfterPreview } from '@/components/analysis/BeforeAfterPreview.tsx'
import { CriterionCard } from '@/components/analysis/CriterionCard.tsx'
import { GenrePanel } from '@/components/analysis/GenrePanel.tsx'
import { StyleCandidateChip } from '@/components/analysis/StyleCandidateChip.tsx'
import { StyleOverview } from '@/components/analysis/StyleOverview.tsx'
import { StyleTab, StyleTabList, StyleTabPanel, StyleTabs } from '@/components/analysis/StyleTabs.tsx'
import { SuggestionPanel } from '@/components/analysis/SuggestionPanel.tsx'
import { Button } from '@/components/ui/Button.tsx'
import { Notice } from '@/components/ui/Notice.tsx'
import { criterionById } from '@/data/criterionProfiles.ts'
import { evaluationSample, genreSample, styleSample } from '@/data/landscapeSample.ts'
import { landscapeStyleProfiles } from '@/data/styleProfiles.ts'
import { matchLevelLabel, styleLabels } from '@/lib/labels.ts'
import type { CriterionEvaluation, CriterionProfile, StyleProfile } from '@/types/analysis.ts'
import type { SelectedPhoto } from '@/types/photo.ts'
import type { SuggestionItem } from '@/types/preview.ts'
import styles from './ResultScreen.module.css'

const DEMO_SUBLINE =
  'Ảnh có màu sắc dịu và ánh sáng khá tự nhiên. Một số tiêu chí bố cục có thể phù hợp với nhiều style.'

interface CriterionView {
  profile: CriterionProfile
  evaluation: CriterionEvaluation
  weight: number
}

interface RubricView {
  profile: StyleProfile
  cards: CriterionView[]
  suggestions: SuggestionItem[]
}

/** Ghép rubric của style với kết quả worker; dữ liệu mẫu, chưa có evaluator thật. */
function rubricFor(styleId: string): RubricView | null {
  const profile = landscapeStyleProfiles.find((item) => item.style_id === styleId)
  const styleEvaluation = evaluationSample.find((item) => item.style_id === styleId)

  if (!profile || !styleEvaluation) return null

  const cards = profile.criteria.flatMap((criterion) => {
    const criterionProfile = criterionById(criterion.criterion_id)
    const evaluation = styleEvaluation.evaluations.find((entry) => entry.criterion_id === criterion.criterion_id)

    if (!criterionProfile || !evaluation) return []

    return [{ profile: criterionProfile, evaluation, weight: criterion.weight }]
  })

  const suggestions = cards.flatMap(({ profile: criterion, evaluation }) =>
    evaluation.suggestion === null
      ? []
      : [
          {
            criterionId: evaluation.criterion_id,
            criterionName: criterion.name,
            text: evaluation.suggestion,
            previewFilter: evaluation.preview_filter ?? null,
          },
        ],
  )

  return { profile, cards, suggestions }
}

interface ResultScreenProps {
  photo: SelectedPhoto
  onChangePhoto: () => void
}

export function ResultScreen({ photo, onChangePhoto }: ResultScreenProps) {
  const [selectedStyle, setSelectedStyle] = useState<string>(styleSample.styles[0]?.name ?? '')
  const [activeSuggestionId, setActiveSuggestionId] = useState<string | null>(null)
  const [skippedIds, setSkippedIds] = useState<string[]>([])
  const [isPhotoExpanded, setIsPhotoExpanded] = useState(false)

  const selectedView = rubricFor(selectedStyle)

  const handleStyleSelect = (value: string) => {
    setSelectedStyle(value)
    setActiveSuggestionId(null)
  }

  const handleSkip = (criterionId: string) => {
    setActiveSuggestionId((current) => (current === criterionId ? null : current))
    setSkippedIds((current) => (current.includes(criterionId) ? current : [...current, criterionId]))
  }

  const handleRestore = (criterionId: string) => {
    setSkippedIds((current) => current.filter((id) => id !== criterionId))
  }

  const renderStyleContent = (styleId: string, isActiveStyle: boolean) => {
    const view = rubricFor(styleId)

    if (!view) {
      return <p className={styles.missing}>Style này chưa có rubric trong dữ liệu mẫu.</p>
    }

    const activeCard = view.cards.find(({ evaluation }) => evaluation.criterion_id === activeSuggestionId)
    const previewFilter = isActiveStyle ? (activeCard?.evaluation.preview_filter ?? null) : null
    const previewLabel = isActiveStyle ? (activeCard?.evaluation.suggestion ?? null) : null

    return (
      <>
        <StyleOverview profile={view.profile} analysis={styleSample} />

        <section aria-labelledby="criteria-heading">
          <h2 className="section-title" id="criteria-heading">
            Tiêu chí theo rubric
          </h2>
          <p className={styles.criteriaNote}>
            Điểm hiển thị riêng theo từng tiêu chí; không gộp các style thành một điểm tổng duy nhất.
          </p>
          <div className={styles.cards}>
            {view.cards.map(({ profile, evaluation, weight }) => (
              <CriterionCard
                key={`${styleId}-${evaluation.criterion_id}`}
                profile={profile}
                evaluation={evaluation}
                weight={weight}
              />
            ))}
          </div>
        </section>

        <SuggestionPanel
          items={view.suggestions}
          activeCriterionId={isActiveStyle ? activeSuggestionId : null}
          skippedCriterionIds={skippedIds}
          onPreview={setActiveSuggestionId}
          onSkip={handleSkip}
          onRestore={handleRestore}
        />

        <BeforeAfterPreview
          photoUrl={photo.url}
          alt={photo.alt}
          previewFilter={previewFilter}
          previewLabel={previewLabel}
        />
      </>
    )
  }

  if (selectedView === null) {
    return (
      <section className={styles.screen}>
        <Notice tone="caution">Chưa có rubric cho style đang chọn. Thử chọn ảnh mẫu để xem dữ liệu minh họa.</Notice>
        <Button variant="ghost" onClick={onChangePhoto}>
          Chọn ảnh khác
        </Button>
      </section>
    )
  }

  const headline = `Bức ảnh có xu hướng ${styleLabels[selectedView.profile.style_id]}`

  return (
    <section className={styles.screen} aria-labelledby="result-heading">
      <div className={styles.topBar}>
        <Button variant="quiet" onClick={onChangePhoto}>
          Chọn ảnh khác
        </Button>
        <p className={styles.topStatus} role="status">
          Hoàn thành lượt phân tích mẫu
        </p>
      </div>

      <div className={styles.layout}>
        <div className={styles.photoCol}>
          <h2 className="section-title">Ảnh gốc</h2>
          <figure className={styles.photo}>
            <img
              className={styles.photoImage}
              data-expanded={isPhotoExpanded}
              src={photo.url}
              alt={photo.alt}
            />
            <figcaption className={styles.photoBar}>
              <span className={styles.fileName}>{photo.name}</span>
              <Button variant="quiet" aria-pressed={isPhotoExpanded} onClick={() => setIsPhotoExpanded((prev) => !prev)}>
                {isPhotoExpanded ? 'Thu nhỏ ảnh' : 'Xem ảnh lớn'}
              </Button>
            </figcaption>
          </figure>

          <GenrePanel classification={genreSample} />
        </div>

        <div className={styles.resultCol}>
          <header className={styles.resultHead}>
            <p className={styles.kicker}>Kết quả theo style</p>
            <h1 id="result-heading">{headline}</h1>
            <p className={styles.subline}>{DEMO_SUBLINE}</p>
          </header>

          <Notice tone="info">
            Dữ liệu mẫu — prototype chưa chạy model hay backend. Ảnh gốc luôn được giữ nguyên.
          </Notice>

          <StyleTabs selected={selectedStyle} onSelect={handleStyleSelect}>
            <StyleTabList label="Style ứng viên cho ảnh này">
              {styleSample.styles.map((candidate, index) => (
                <StyleTab key={candidate.name} value={candidate.name}>
                  <StyleCandidateChip
                    name={styleLabels[candidate.name]}
                    matchLabel={matchLevelLabel(index, styleSample.styles.length)}
                    score={candidate.score}
                  />
                </StyleTab>
              ))}
            </StyleTabList>

            {styleSample.styles.map((candidate) => (
              <StyleTabPanel key={candidate.name} value={candidate.name}>
                {renderStyleContent(candidate.name, candidate.name === selectedStyle)}
              </StyleTabPanel>
            ))}
          </StyleTabs>
        </div>
      </div>
    </section>
  )
}
