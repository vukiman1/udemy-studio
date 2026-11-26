import { isCompressing } from '../../api/compression-queries'
import { isPublishing } from '../../api/publish-queries'
import { isTranslating } from '../../api/translation-queries'
import type { CourseTools, PublishRun } from '../../api/types'

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

const TRANSLATE_LABEL = 'Translate'
const COMPRESS_LABEL = 'Compress'
const PUBLISH_LABEL = 'Publish'
const PUBLISHED_LABEL = 'Update R2'
const PERCENT = 100

const NOTHING_RUNNING = 0

// "Done" comes from the library, not from the last run, so lectures downloaded later make the tool available again.
export function translationStatus(tools: CourseTools['translation']): ToolStatus {
  const { run, pending, lectures } = tools
  if (run !== undefined && isTranslating(run))
    return {
      state: ToolState.Running,
      progress: run.total === 0 ? 0 : run.finished / run.total,
      label: `Translating ${String(run.finished)}/${String(run.total)}`,
      detail: `Translating ${String(run.finished)}/${String(run.total)} lectures`,
    }
  if (lectures === 0)
    return {
      state: ToolState.Unavailable,
      progress: NOTHING_RUNNING,
      label: TRANSLATE_LABEL,
      detail: 'No English subtitles',
    }
  if (pending === 0)
    return {
      state: ToolState.Done,
      progress: NOTHING_RUNNING,
      label: TRANSLATE_LABEL,
      detail: `All ${String(lectures)} lectures translated`,
    }
  if (run !== undefined && (run.state === 'failed' || run.failed.length > 0))
    return {
      state: ToolState.Failed,
      progress: NOTHING_RUNNING,
      label: TRANSLATE_LABEL,
      detail: `Last run had errors; ${String(pending)} left`,
    }
  return {
    state: ToolState.Idle,
    progress: NOTHING_RUNNING,
    label: TRANSLATE_LABEL,
    detail: `${String(pending)}/${String(lectures)} lectures to translate`,
  }
}

export function compressionStatus(tools: CourseTools['compression']): ToolStatus {
  if (tools === undefined)
    return {
      state: ToolState.Unavailable,
      progress: NOTHING_RUNNING,
      label: COMPRESS_LABEL,
      detail: 'ffmpeg not installed',
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
      label: `Compressing ${String(Math.min(run.finished + 1, run.total))}/${String(run.total)}`,
      detail: `Compressing ${String(Math.min(run.finished + 1, run.total))}/${String(run.total)} videos`,
    }
  }
  if (unchecked > 0 && pending === 0)
    return {
      state: ToolState.Checking,
      progress: NOTHING_RUNNING,
      label: `Checking ${String(videos - unchecked)}/${String(videos)}`,
      detail: `Checking videos (${String(videos - unchecked)}/${String(videos)})`,
    }
  if (pending === 0)
    return {
      state: ToolState.Done,
      progress: NOTHING_RUNNING,
      label: COMPRESS_LABEL,
      detail: 'Nothing left to compress',
    }
  if (run !== undefined && run.failed.length > 0)
    return {
      state: ToolState.Failed,
      progress: NOTHING_RUNNING,
      label: COMPRESS_LABEL,
      detail: `Last run had errors; ${String(pending)} H.264 videos left`,
    }
  return {
    state: ToolState.Idle,
    progress: NOTHING_RUNNING,
    label: COMPRESS_LABEL,
    detail: `${String(pending)} H.264 videos to compress`,
  }
}

export function publishStatus(run: PublishRun | undefined, isPublished: boolean): ToolStatus {
  if (run !== undefined && isPublishing(run)) {
    const share = run.totalBytes === 0 ? 0 : run.sentBytes / run.totalBytes
    return {
      state: ToolState.Running,
      progress: share,
      label: `Publishing ${String(Math.round(share * PERCENT))}%`,
      detail: `Publishing ${String(run.sentFiles + run.skippedFiles)}/${String(run.totalFiles)} files`,
    }
  }
  if (run?.state === 'failed')
    return {
      state: ToolState.Failed,
      progress: NOTHING_RUNNING,
      label: PUBLISH_LABEL,
      detail: 'Last publish failed',
    }
  if (isPublished)
    return {
      state: ToolState.Idle,
      progress: NOTHING_RUNNING,
      label: PUBLISHED_LABEL,
      detail: 'On R2; send new or changed files',
    }
  return {
    state: ToolState.Idle,
    progress: NOTHING_RUNNING,
    label: PUBLISH_LABEL,
    detail: 'Publish to R2',
  }
}
