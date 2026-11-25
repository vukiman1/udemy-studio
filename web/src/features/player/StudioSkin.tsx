import {
  Container,
  Controls,
  ErrorDialog,
  FullscreenButton,
  Gesture,
  Hotkey,
  MuteButton,
  PlayButton,
  StatusAnnouncer,
  Time,
  TimeSlider,
  VolumeSlider,
} from '@videojs/react'
import {
  CaptionsOffIcon,
  CaptionsOnIcon,
  FullscreenEnterIcon,
  FullscreenExitIcon,
  PauseIcon,
  PlayIcon,
  VolumeHighIcon,
  VolumeOffIcon,
} from '@videojs/react/icons'
import type { ReactElement, ReactNode } from 'react'
import { PlayerCenterState } from './PlayerCenterState'

const SHORT_SEEK_SECONDS = 5
const LONG_SEEK_SECONDS = 10

const CONTROL_BUTTON =
  'inline-grid size-10 place-items-center rounded-md border-0 bg-transparent p-0 text-inherit hover:bg-white/15 focus-visible:bg-white/15 aria-pressed:text-accent'
const CONTROL_ICON = 'size-[22px]'
const SLIDER_ROOT =
  'group/slider pointer-events-auto relative flex h-[18px] cursor-pointer touch-none items-center'
const SLIDER_TRACK =
  'relative h-1 flex-1 overflow-hidden rounded-sm bg-white/25 group-hover/slider:h-1.5'
const SLIDER_FILL = 'absolute inset-y-0 left-0 w-[var(--media-slider-fill,0%)] bg-accent'
const SLIDER_THUMB =
  'absolute left-[var(--media-slider-fill,0%)] size-3.5 -translate-x-1/2 rounded-full bg-white'

interface StudioSkinProps {
  readonly children: ReactNode
  readonly overlay: ReactNode
  readonly settingsMenu: ReactNode
  readonly areSubtitlesOn: boolean
  readonly onToggleSubtitles: () => void
}

export function StudioSkin(props: StudioSkinProps): ReactElement {
  return (
    <Container className="@container relative block overflow-hidden rounded-2xl bg-black text-white [--subtitle-offset:6%] has-[[data-slot=controls][data-visible]]:[--subtitle-offset:76px] [&_video]:block [&_video]:aspect-video [&_video]:max-h-[75dvh] [&_video]:w-full [&_video]:bg-black [&:fullscreen]:rounded-none [&:fullscreen_video]:h-dvh [&:fullscreen_video]:max-h-none">
      {props.children}
      {props.overlay}
      <PlayerCenterState />

      <Controls.Root>
        <Controls.Backdrop className="pointer-events-none absolute inset-x-0 bottom-0 h-2/5 bg-linear-to-t from-black/70 to-transparent transition-opacity duration-300 not-data-visible:opacity-0" />
        <Controls.Content
          data-slot="controls"
          className="pointer-events-none absolute inset-x-0 bottom-0 px-3 pb-1.5 transition-opacity duration-250 not-data-visible:invisible not-data-visible:opacity-0"
        >
          <TimeSlider.Root className={SLIDER_ROOT}>
            <TimeSlider.Track className={SLIDER_TRACK}>
              <TimeSlider.Buffer className="absolute inset-y-0 left-0 w-[var(--media-slider-buffer,0%)] bg-white/35" />
              <TimeSlider.Fill className={SLIDER_FILL} />
            </TimeSlider.Track>
            <TimeSlider.Thumb className={SLIDER_THUMB} />
          </TimeSlider.Root>

          <Controls.Group
            className="pointer-events-auto flex items-center gap-0.5"
            aria-label="Điều khiển phát"
          >
            <PlayButton
              className={CONTROL_BUTTON}
              render={(attributes, state) => (
                <button {...attributes}>
                  {state.paused ? (
                    <PlayIcon className={CONTROL_ICON} />
                  ) : (
                    <PauseIcon className={CONTROL_ICON} />
                  )}
                </button>
              )}
            />
            <MuteButton
              className={CONTROL_BUTTON}
              render={(attributes, state) => (
                <button {...attributes}>
                  {state.muted ? (
                    <VolumeOffIcon className={CONTROL_ICON} />
                  ) : (
                    <VolumeHighIcon className={CONTROL_ICON} />
                  )}
                </button>
              )}
            />
            <VolumeSlider.Root className={`${SLIDER_ROOT} w-20 pointer-coarse:hidden`}>
              <VolumeSlider.Track className={SLIDER_TRACK}>
                <VolumeSlider.Fill className={SLIDER_FILL} />
              </VolumeSlider.Track>
              <VolumeSlider.Thumb className={SLIDER_THUMB} />
            </VolumeSlider.Root>
            <Time.Group className="ml-2 inline-flex gap-1 text-[0.85rem] tabular-nums">
              <Time.Value type="current" />
              <Time.Separator />
              <Time.Value type="duration" />
            </Time.Group>
            <span className="flex-1" />
            <button
              type="button"
              className={CONTROL_BUTTON}
              aria-pressed={props.areSubtitlesOn}
              aria-label={props.areSubtitlesOn ? 'Tắt phụ đề' : 'Bật phụ đề'}
              onClick={props.onToggleSubtitles}
            >
              {props.areSubtitlesOn ? (
                <CaptionsOnIcon className={CONTROL_ICON} />
              ) : (
                <CaptionsOffIcon className={CONTROL_ICON} />
              )}
            </button>
            {props.settingsMenu}
            <FullscreenButton
              className={CONTROL_BUTTON}
              render={(attributes, state) => (
                <button {...attributes}>
                  {state.fullscreen ? (
                    <FullscreenExitIcon className={CONTROL_ICON} />
                  ) : (
                    <FullscreenEnterIcon className={CONTROL_ICON} />
                  )}
                </button>
              )}
            />
          </Controls.Group>
        </Controls.Content>
      </Controls.Root>

      <Hotkey keys="Space" action="togglePaused" target="document" />
      <Hotkey keys="k" action="togglePaused" target="document" />
      <Hotkey keys="m" action="toggleMuted" target="document" />
      <Hotkey keys="f" action="toggleFullscreen" target="document" />
      <Hotkey keys="ArrowLeft" action="seekStep" value={-SHORT_SEEK_SECONDS} target="document" />
      <Hotkey keys="ArrowRight" action="seekStep" value={SHORT_SEEK_SECONDS} target="document" />
      <Hotkey keys="j" action="seekStep" value={-LONG_SEEK_SECONDS} target="document" />
      <Hotkey keys="l" action="seekStep" value={LONG_SEEK_SECONDS} target="document" />
      <Gesture type="tap" pointer="mouse" action="togglePaused" />
      <Gesture type="tap" pointer="touch" action="toggleControls" />
      <Gesture type="doubletap" pointer="mouse" action="toggleFullscreen" />
      <Gesture
        type="doubletap"
        pointer="touch"
        region="left"
        action="seekStep"
        value={-LONG_SEEK_SECONDS}
      />
      <Gesture
        type="doubletap"
        pointer="touch"
        region="right"
        action="seekStep"
        value={LONG_SEEK_SECONDS}
      />

      <StatusAnnouncer className="pointer-events-none absolute top-1/2 left-1/2 mt-12 -translate-x-1/2 rounded-full bg-black/60 px-3 py-1 text-[13px] whitespace-nowrap backdrop-blur-sm has-[[data-status-announcer-content]:empty]:hidden" />
      <ErrorDialog.Root>
        <ErrorDialog.Popup className="absolute inset-0 grid place-content-center gap-2 bg-black/80 p-6 text-center">
          <ErrorDialog.Title />
          <ErrorDialog.Description />
          <ErrorDialog.Close />
        </ErrorDialog.Popup>
      </ErrorDialog.Root>
    </Container>
  )
}
