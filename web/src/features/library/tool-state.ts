import { isCompressing } from '../../api/compression-queries'
import { isTranslating } from '../../api/translation-queries'
import type { CourseTools } from '../../api/types'

export enum ToolState {
  Idle = 'idle',
  Running = 'running',
  Done = 'done',
  Failed = 'failed',
  Checking = 'checking',
  Unavailable = 'unavailable',
}

export interface ToolStatus {
  readonly state: ToolState
  // Share of the running job done, from 0 to 1; zero when nothing runs.
  readonly progress: number
  readonly label: string
  readonly detail: string
}

const TRANSLATE_LABEL = 'Dịch'
const COMPRESS_LABEL = 'Nén'

const NOTHING_RUNNING = 0

// "Done" comes from the library, not from the last run, so lectures downloaded later make the tool available again.
export function translationStatus(tools: CourseTools['translation']): ToolStatus {
  const { run, pending, lectures } = tools
  if (run !== undefined && isTranslating(run))
    return {
      state: ToolState.Running,
      progress: run.total === 0 ? 0 : run.finished / run.total,
      label: `Đang dịch ${String(run.finished)}/${String(run.total)}`,
      detail: `Đang dịch ${String(run.finished)}/${String(run.total)} bài`,
    }
  if (lectures === 0)
    return {
      state: ToolState.Unavailable,
      progress: NOTHING_RUNNING,
      label: TRANSLATE_LABEL,
      detail: 'Khoá này không có phụ đề tiếng Anh',
    }
  if (pending === 0)
    return {
      state: ToolState.Done,
      progress: NOTHING_RUNNING,
      label: TRANSLATE_LABEL,
      detail: `Đã dịch đủ ${String(lectures)}/${String(lectures)} bài`,
    }
  if (run !== undefined && (run.state === 'failed' || run.failed.length > 0))
    return {
      state: ToolState.Failed,
      progress: NOTHING_RUNNING,
      label: TRANSLATE_LABEL,
      detail: `Lần dịch trước có lỗi; còn ${String(pending)} bài chưa dịch`,
    }
  return {
    state: ToolState.Idle,
    progress: NOTHING_RUNNING,
    label: TRANSLATE_LABEL,
    detail: `Dịch phụ đề: còn ${String(pending)}/${String(lectures)} bài`,
  }
}

export function compressionStatus(tools: CourseTools['compression']): ToolStatus {
  if (tools === undefined)
    return {
      state: ToolState.Unavailable,
      progress: NOTHING_RUNNING,
      label: COMPRESS_LABEL,
      detail: 'Máy chưa có ffmpeg',
    }
  const { run, pending, unchecked, videos } = tools
  if (run !== undefined && isCompressing(run)) {
    const current = run.current
    const currentShare =
      current === undefined || current.lengthSeconds === 0
        ? 0
        : current.encodedSeconds / current.lengthSeconds
    return {
      state: ToolState.Running,
      progress: run.total === 0 ? 0 : (run.finished + Math.min(1, currentShare)) / run.total,
      label: `Đang nén ${String(Math.min(run.finished + 1, run.total))}/${String(run.total)}`,
      detail: `Đang nén ${String(Math.min(run.finished + 1, run.total))}/${String(run.total)} video`,
    }
  }
  if (unchecked > 0 && pending === 0)
    return {
      state: ToolState.Checking,
      progress: NOTHING_RUNNING,
      label: `Kiểm tra ${String(videos - unchecked)}/${String(videos)}`,
      detail: `Đang kiểm tra video (${String(videos - unchecked)}/${String(videos)})`,
    }
  if (pending === 0)
    return {
      state: ToolState.Done,
      progress: NOTHING_RUNNING,
      label: COMPRESS_LABEL,
      detail: 'Không còn video nào cần nén',
    }
  if (run !== undefined && run.failed.length > 0)
    return {
      state: ToolState.Failed,
      progress: NOTHING_RUNNING,
      label: COMPRESS_LABEL,
      detail: `Lần nén trước có lỗi; còn ${String(pending)} video H.264`,
    }
  return {
    state: ToolState.Idle,
    progress: NOTHING_RUNNING,
    label: COMPRESS_LABEL,
    detail: `Nén video: còn ${String(pending)} video H.264`,
  }
}
