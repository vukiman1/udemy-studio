import { Menu } from '@videojs/react'
import { CheckIcon, ChevronIcon, GearIcon } from '@videojs/react/icons'
import type { ReactElement } from 'react'
import {
  FONT_SCALES,
  PLAYBACK_RATES,
  SubtitleMode,
  parseSubtitleMode,
  type PlayerSettings,
} from './player-settings'
import styles from './SettingsMenu.module.css'

interface SettingsMenuProps {
  readonly settings: PlayerSettings
  readonly subtitleModes: readonly SubtitleMode[]
  readonly onChange: (update: (current: PlayerSettings) => PlayerSettings) => void
}

interface Choice {
  readonly value: string
  readonly label: string
}

const PERCENT = 100

export const SUBTITLE_MODE_LABELS: Readonly<Record<SubtitleMode, string>> = {
  [SubtitleMode.Bilingual]: 'Song ngữ',
  [SubtitleMode.Primary]: 'Tiếng Việt',
  [SubtitleMode.Secondary]: 'English',
  [SubtitleMode.Off]: 'Tắt',
}

export function SettingsMenu({
  settings,
  subtitleModes,
  onChange,
}: SettingsMenuProps): ReactElement {
  return (
    <Menu.Root side="top" align="end">
      <Menu.Trigger
        className={styles.trigger}
        aria-label="Cài đặt"
        render={<button type="button" />}
      >
        <GearIcon className={styles.icon} />
      </Menu.Trigger>
      <Menu.Popup className={styles.popup}>
        <Menu.Content className={styles.content}>
          <SettingsSubmenu
            label="Tốc độ"
            value={String(settings.playbackRate)}
            choices={PLAYBACK_RATES.map((rate) => ({
              value: String(rate),
              label: `${String(rate)}×`,
            }))}
            onSelect={(value) => {
              onChange((current) => ({ ...current, playbackRate: Number(value) }))
            }}
          />
          {subtitleModes.length > 1 && (
            <SettingsSubmenu
              label="Phụ đề"
              value={settings.subtitleMode}
              choices={subtitleModes.map((mode) => ({
                value: mode,
                label: SUBTITLE_MODE_LABELS[mode],
              }))}
              onSelect={(value) => {
                const subtitleMode = parseSubtitleMode(value)
                if (subtitleMode !== undefined)
                  onChange((current) => ({ ...current, subtitleMode }))
              }}
            />
          )}
          <SettingsSubmenu
            label="Cỡ chữ phụ đề"
            value={String(settings.fontScale)}
            choices={FONT_SCALES.map((scale) => ({
              value: String(scale),
              label: `${String(Math.round(scale * PERCENT))}%`,
            }))}
            onSelect={(value) => {
              onChange((current) => ({ ...current, fontScale: Number(value) }))
            }}
          />
        </Menu.Content>
      </Menu.Popup>
    </Menu.Root>
  )
}

interface SettingsSubmenuProps {
  readonly label: string
  readonly value: string
  readonly choices: readonly Choice[]
  readonly onSelect: (value: string) => void
}

function SettingsSubmenu({ label, value, choices, onSelect }: SettingsSubmenuProps): ReactElement {
  const selected = choices.find((choice) => choice.value === value)
  return (
    <Menu.Root>
      <Menu.Trigger
        className={styles.item}
        render={(props) => (
          <div {...props}>
            <span>{label}</span>
            <span className={styles.value}>
              {selected?.label}
              <ChevronIcon className={styles.chevron} />
            </span>
          </div>
        )}
      />
      <Menu.Content className={styles.content}>
        <Menu.Item className={styles.back}>
          <ChevronIcon className={styles.backChevron} />
          {label}
        </Menu.Item>
        <Menu.RadioGroup value={value} onValueChange={onSelect} aria-label={label}>
          {choices.map((choice) => (
            <Menu.RadioItem key={choice.value} value={choice.value} className={styles.item}>
              <span>{choice.label}</span>
              <Menu.ItemIndicator checked={choice.value === value} className={styles.indicator}>
                <CheckIcon className={styles.icon} />
              </Menu.ItemIndicator>
            </Menu.RadioItem>
          ))}
        </Menu.RadioGroup>
      </Menu.Content>
    </Menu.Root>
  )
}
