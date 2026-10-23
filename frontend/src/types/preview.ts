export interface SuggestionItem {
  criterionId: string
  criterionName: string
  text: string
  /** prototype only: CSS filter mô phỏng, chưa có hợp đồng backend */
  previewFilter: string | null
}
