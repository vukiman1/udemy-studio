import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { selectChromeProfile, useChromeProfiles, useLoginStatus } from '../../api/download-queries'
import { QUERY_KEYS } from '../../api/queries'
import { TEXT_FIELD } from '../../ui/text-field'
import { DISPLAY_LOCALE } from '../../app/display-locale'

const DATE_FORMAT = new Intl.DateTimeFormat(DISPLAY_LOCALE, { dateStyle: 'medium' })

// A machine without Chrome, such as a Docker host, lists no profiles, so the picker gives way to pasted tokens.
export function ProfilePicker(): ReactElement | null {
  const queryClient = useQueryClient()
  const profiles = useChromeProfiles()
  const status = useLoginStatus()
  const select = useMutation({
    mutationFn: selectChromeProfile,
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.chromeProfiles }),
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.loginStatus }),
      ])
    },
  })
  const selected = profiles.data?.find((profile) => profile.selected)
  if (profiles.data?.length === 0) return null

  return (
    <div className="grid gap-4 border-b border-line pb-4">
      <label className="grid gap-2 text-sm font-semibold">
        Chrome profile login
        <select
          className={`${TEXT_FIELD} font-normal`}
          value={selected?.dir ?? ''}
          disabled={profiles.data === undefined || select.isPending}
          onChange={(event) => {
            select.mutate(event.target.value)
          }}
        >
          <option value="" disabled>
            Choose a profile…
          </option>
          {profiles.data?.map((profile) => (
            <option key={profile.dir} value={profile.dir}>
              {profile.name} ({profile.dir})
              {profile.udemyHosts.length > 0 ? ` — ${profile.udemyHosts.join(', ')}` : ''}
            </option>
          ))}
        </select>
      </label>
      {profiles.isError && (
        <p className="text-[13px] text-danger" role="alert">
          {profiles.error.message}
        </p>
      )}
      {select.isError && (
        <p className="text-[13px] text-danger" role="alert">
          {select.error.message}
        </p>
      )}
      {status.data?.problem !== undefined && (
        <p className="text-[13px] text-danger" role="alert">
          {status.data.problem}
        </p>
      )}
      {status.data !== undefined && status.data.hosts.length > 0 && (
        <ul className="grid list-none gap-1.5 text-[13px]">
          {status.data.hosts.map((host) => (
            <li
              key={host.host}
              className={`flex items-start gap-2 ${host.isExpired ? 'text-danger' : 'text-muted'}`}
            >
              <span
                className={`mt-[5px] size-2 shrink-0 rounded-full ${host.isExpired ? 'bg-danger' : 'bg-success'}`}
                aria-hidden="true"
              />
              {host.host} —{' '}
              {host.isExpired
                ? 'login expired; sign in again in Chrome'
                : `logged in${host.expiresAt === undefined ? '' : ` until ${DATE_FORMAT.format(new Date(host.expiresAt))}`}`}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
