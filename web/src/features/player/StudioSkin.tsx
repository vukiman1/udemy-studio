import {
  BufferingIndicator,
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
  SpinnerIcon,
  VolumeHighIcon,
  VolumeOffIcon,
} from '@videojs/react/icons'
import type { ReactElement, ReactNode } from 'react'
import styles from './StudioSkin.module.css'

const SHORT_SEEK_SECONDS = 5
const LONG_SEEK_SECONDS = 10

interface StudioSkinProps {
  readonly children: ReactNode
  readonly overlay: ReactNode
  readonly settingsMenu: ReactNode
  readonly areSubtitlesOn: boolean
  readonly onToggleSubtitles: () => void
}

export function StudioSkin(props: StudioSkinProps): ReactElement {
  return (
    <Container className={styles.skin}>
      {props.children}
      {props.overlay}
      <BufferingIndicator
        className={styles.buffering}
        render={(attributes) => (
          <div {...attributes}>
            <SpinnerIcon className={styles.spinner} />
          </div>
        )}
      />

      <Controls.Root>
        <Controls.Backdrop className={styles.backdrop} />
        <Controls.Content className={styles.controls}>
          <TimeSlider.Root className={styles.timeSlider}>
            <TimeSlider.Track className={styles.track}>
              <TimeSlider.Buffer className={styles.buffer} />
              <TimeSlider.Fill className={styles.fill} />
            </TimeSlider.Track>
            <TimeSlider.Thumb className={styles.thumb} />
          </TimeSlider.Root>

          <Controls.Group className={styles.bar} aria-label="Điều khiển phát">
            <PlayButton
              className={styles.button}
              render={(attributes, state) => (
                <button {...attributes}>
                  {state.paused ? (
                    <PlayIcon className={styles.icon} />
                  ) : (
                    <PauseIcon className={styles.icon} />
                  )}
                </button>
              )}
            />
            <MuteButton
              className={styles.button}
              render={(attributes, state) => (
                <button {...attributes}>
                  {state.muted ? (
                    <VolumeOffIcon className={styles.icon} />
                  ) : (
                    <VolumeHighIcon className={styles.icon} />
                  )}
                </button>
              )}
            />
            <VolumeSlider.Root className={styles.volume}>
              <VolumeSlider.Track className={styles.track}>
                <VolumeSlider.Fill className={styles.fill} />
              </VolumeSlider.Track>
              <VolumeSlider.Thumb className={styles.thumb} />
            </VolumeSlider.Root>
            <Time.Group className={styles.time}>
              <Time.Value type="current" />
              <Time.Separator />
              <Time.Value type="duration" />
            </Time.Group>
            <span className={styles.spacer} />
            <button
              type="button"
              className={styles.button}
              aria-pressed={props.areSubtitlesOn}
              aria-label={props.areSubtitlesOn ? 'Tắt phụ đề' : 'Bật phụ đề'}
              onClick={props.onToggleSubtitles}
            >
              {props.areSubtitlesOn ? (
                <CaptionsOnIcon className={styles.icon} />
              ) : (
                <CaptionsOffIcon className={styles.icon} />
              )}
            </button>
            {props.settingsMenu}
            <FullscreenButton
              className={styles.button}
              render={(attributes, state) => (
                <button {...attributes}>
                  {state.fullscreen ? (
                    <FullscreenExitIcon className={styles.icon} />
                  ) : (
                    <FullscreenEnterIcon className={styles.icon} />
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

      <StatusAnnouncer />
      <ErrorDialog.Root>
        <ErrorDialog.Popup className={styles.error}>
          <ErrorDialog.Title />
          <ErrorDialog.Description />
          <ErrorDialog.Close />
        </ErrorDialog.Popup>
      </ErrorDialog.Root>
    </Container>
  )
}
