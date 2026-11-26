import { Menu } from '@videojs/react'
import { CheckIcon, ChevronIcon, GearIcon } from '@videojs/react/icons'
import type { ReactElement } from 'react'
import {
  BACKGROUND_OPACITIES,
  FONT_SCALES,
  PLAYBACK_RATES,
  TEXT_OPACITIES,
  SubtitleMode,
  parseSubtitleMode,
  type PlayerSettings,
} from './player-settings'

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

const MENU_PAGE =
  'col-start-1 row-start-1 grid content-start outline-none data-child-open:invisible'
const MENU_ITEM =
  'flex cursor-pointer items-center justify-between gap-4 rounded-md px-2.5 py-2 outline-none data-highlighted:bg-white/12'

export const SUBTITLE_MODE_LABELS: Readonly<Record<SubtitleMode, string>> = {
  [SubtitleMode.Bilingual]: 'Bilingual',
  [SubtitleMode.Primary]: 'Vietnamese',
  [SubtitleMode.Secondary]: 'English',
  [SubtitleMode.Off]: 'Off',
}

export function SettingsMenu({
  settings,
  subtitleModes,
  onChange,
}: SettingsMenuProps): ReactElement {
  return (
    <Menu.Root side="top" align="end">
      <Menu.Trigger
        className="inline-grid size-10 place-items-center rounded-md border-0 bg-transparent p-0 text-inherit hover:bg-white/15 data-open:bg-white/15"
        aria-label="Settings"
        render={<button type="button" />}
      >
        <GearIcon className="size-5" />
      </Menu.Trigger>
      <Menu.Popup className="grid min-w-[230px] rounded-[10px] bg-[rgb(28_28_32/0.96)] p-1.5 text-[0.9rem] text-white shadow-[0_8px_24px_rgb(0_0_0/0.4)] outline-none">
        <Menu.Content className={MENU_PAGE}>
          <SettingsSubmenu
            label="Speed"
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
              label="Subtitles"
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
            label="Subtitle size"
            value={String(settings.fontScale)}
            choices={FONT_SCALES.map((scale) => ({
              value: String(scale),
              label: asPercent(scale),
            }))}
            onSelect={(value) => {
              onChange((current) => ({ ...current, fontScale: Number(value) }))
            }}
          />
          <SettingsSubmenu
            label="Subtitle opacity"
            value={String(settings.textOpacity)}
            choices={TEXT_OPACITIES.map((opacity) => ({
              value: String(opacity),
              label: asPercent(opacity),
            }))}
            onSelect={(value) => {
              onChange((current) => ({ ...current, textOpacity: Number(value) }))
            }}
          />
          <SettingsSubmenu
            label="Subtitle background"
            value={String(settings.backgroundOpacity)}
            choices={BACKGROUND_OPACITIES.map((opacity) => ({
              value: String(opacity),
              label: opacity === 0 ? 'None' : asPercent(opacity),
            }))}
            onSelect={(value) => {
              onChange((current) => ({ ...current, backgroundOpacity: Number(value) }))
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
        className={MENU_ITEM}
        render={(props) => (
          <div {...props}>
            <span>{label}</span>
            <span className="inline-flex items-center gap-1 text-white/70">
              {selected?.label}
              <ChevronIcon className="size-3.5" />
            </span>
          </div>
        )}
      />
      <Menu.Content className={MENU_PAGE}>
        <Menu.Item className="mb-1 flex cursor-pointer items-center gap-1.5 rounded-t-md border-b border-white/12 px-2.5 py-2 font-semibold outline-none data-highlighted:bg-white/12">
          <ChevronIcon className="size-3.5 -scale-x-100" />
          {label}
        </Menu.Item>
        <Menu.RadioGroup
          className="grid max-h-[min(16rem,36cqw)] overflow-y-auto"
          value={value}
          onValueChange={onSelect}
          aria-label={label}
        >
          {choices.map((choice) => (
            <Menu.RadioItem key={choice.value} value={choice.value} className={MENU_ITEM}>
              <span>{choice.label}</span>
              <Menu.ItemIndicator checked={choice.value === value} className="inline-grid">
                <CheckIcon className="size-5" />
              </Menu.ItemIndicator>
            </Menu.RadioItem>
          ))}
        </Menu.RadioGroup>
      </Menu.Content>
    </Menu.Root>
  )
}

function asPercent(fraction: number): string {
  return `${String(Math.round(fraction * PERCENT))}%`
}
