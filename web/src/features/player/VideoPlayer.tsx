import { I18nProvider } from '@videojs/react/i18n'
import { Video } from '@videojs/react/video'
import { useEffect, type ReactElement, type RefObject } from 'react'
import { subtitlesVttUrl, useCues, videoUrl } from '../../api/queries'
import type { Cue, LectureSummary } from '../../api/types'
import { Player } from './player'
import {
  FONT_SCALES,
  PLAYBACK_RATES,
  SubtitleMode,
  languagesFor,
  nextSubtitleMode,
  stepThrough,
  toggleSubtitles,
} from './player-settings'
import { SettingsMenu } from './SettingsMenu'
import { StudioSkin } from './StudioSkin'
import { SubtitleOverlay } from './SubtitleOverlay'
import { useActiveLines } from './use-active-lines'
import { useLectureProgress } from './use-lecture-progress'
import { useNativeFullscreenSubtitles } from './use-native-fullscreen-subtitles'
import { usePlayerSettings } from './use-player-settings'
import { useSettingsShortcuts } from './use-settings-shortcuts'

interface VideoPlayerProps {
  readonly videoRef: RefObject<HTMLVideoElement | null>
  readonly lecture: LectureSummary
  readonly courseId: string
  readonly startAtSeconds: number | undefined
}

const NO_CUES: readonly Cue[] = []
const PLAYER_LOCALE = 'vi'

export function VideoPlayer({
  videoRef,
  lecture,
  courseId,
  startAtSeconds,
}: VideoPlayerProps): ReactElement {
  const [settings, updateSettings] = usePlayerSettings()

  const languages = languagesFor(settings.subtitleMode, lecture.subtitleLanguages)
  const cuesQuery = useCues(lecture.id, languages)
  const activeLines = useActiveLines(videoRef, cuesQuery.data ?? NO_CUES, languages)
  const { lastError } = useLectureProgress(videoRef, lecture, courseId, startAtSeconds)
  const trackSource = languages.length > 0 ? subtitlesVttUrl(lecture.id, languages) : undefined
  useNativeFullscreenSubtitles(videoRef, trackSource, languages[0])

  useEffect(() => {
    const video = videoRef.current
    if (video === null) return
    video.preservesPitch = true
    video.defaultPlaybackRate = settings.playbackRate
    video.playbackRate = settings.playbackRate
  }, [videoRef, settings.playbackRate])

  useSettingsShortcuts({
    cycleSubtitles: () => {
      updateSettings((current) => ({
        ...current,
        subtitleMode: nextSubtitleMode(current.subtitleMode),
      }))
    },
    stepRate: (direction) => {
      updateSettings((current) => ({
        ...current,
        playbackRate: stepThrough(PLAYBACK_RATES, current.playbackRate, direction),
      }))
    },
    stepFontScale: (direction) => {
      updateSettings((current) => ({
        ...current,
        fontScale: stepThrough(FONT_SCALES, current.fontScale, direction),
      }))
    },
  })

  const subtitleModes = Object.values(SubtitleMode).filter(
    (mode) => mode === SubtitleMode.Off || languagesFor(mode, lecture.subtitleLanguages).length > 0,
  )
  const areSubtitlesOn = settings.subtitleMode !== SubtitleMode.Off

  return (
    <div className="grid gap-2">
      <Player>
        <I18nProvider locale={PLAYER_LOCALE}>
          <StudioSkin
            overlay={
              <SubtitleOverlay
                lines={activeLines}
                fontScale={settings.fontScale}
                textOpacity={settings.textOpacity}
                backgroundOpacity={settings.backgroundOpacity}
              />
            }
            settingsMenu={
              <SettingsMenu
                settings={settings}
                subtitleModes={subtitleModes}
                onChange={updateSettings}
              />
            }
            areSubtitlesOn={areSubtitlesOn}
            onToggleSubtitles={() => {
              updateSettings((current) => ({
                ...current,
                subtitleMode: toggleSubtitles(current.subtitleMode),
              }))
            }}
          >
            <Video ref={videoRef} src={videoUrl(lecture.id)} playsInline preload="metadata" />
          </StudioSkin>
        </I18nProvider>
      </Player>
      {cuesQuery.isError && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          Không tải được phụ đề ({cuesQuery.error.message}).
        </p>
      )}
      {lastError !== undefined && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          Chưa lưu được tiến độ ({lastError.message}). Sẽ thử lại ở lần lưu sau.
        </p>
      )}
    </div>
  )
}
