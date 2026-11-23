import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { selectChromeProfile, useChromeProfiles, useLoginStatus } from '../../api/download-queries'
import { QUERY_KEYS } from '../../api/queries'

const DATE_FORMAT = new Intl.DateTimeFormat('vi-VN', { dateStyle: 'medium' })

export function ProfilePicker(): ReactElement {
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

  return (
    <section className="grid gap-3 rounded-xl border border-line bg-surface p-4">
      <h2 className="font-semibold">Tài khoản Udemy</h2>
      <label className="grid gap-1 text-[0.9rem]">
        Đọc đăng nhập từ profile Chrome
        <select
          value={selected?.dir ?? ''}
          disabled={profiles.data === undefined || select.isPending}
          onChange={(event) => {
            select.mutate(event.target.value)
          }}
        >
          <option value="" disabled>
            Chọn profile…
          </option>
          {profiles.data?.map((profile) => (
            <option key={profile.dir} value={profile.dir}>
              {profile.name} ({profile.dir})
              {profile.udemyHosts.length > 0 ? ` — ${profile.udemyHosts.join(', ')}` : ''}
            </option>
          ))}
        </select>
      </label>
      {profiles.isError && <p className="text-danger">{profiles.error.message}</p>}
      {select.isError && <p className="text-danger">{select.error.message}</p>}
      {status.data?.problem !== undefined && <p className="text-danger">{status.data.problem}</p>}
      {status.data !== undefined && status.data.hosts.length > 0 && (
        <ul className="grid gap-1 text-[0.9rem]">
          {status.data.hosts.map((host) => (
            <li key={host.host} className={host.isExpired ? 'text-danger' : 'text-muted'}>
              {host.host}:{' '}
              {host.isExpired
                ? 'đăng nhập đã hết hạn, hãy đăng nhập lại trong Chrome'
                : `đăng nhập còn hạn${host.expiresAt === undefined ? '' : ` tới ${DATE_FORMAT.format(new Date(host.expiresAt))}`}`}
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
