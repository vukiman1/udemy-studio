import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { Save } from 'lucide-react'
import { forgetPastedLogin, savePastedLogin, usePastedLogins } from '../../api/download-queries'
import { QUERY_KEYS } from '../../api/queries'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'
import { TEXT_FIELD } from '../../ui/text-field'
import { DISPLAY_LOCALE } from '../../app/display-locale'

const DATE_FORMAT = new Intl.DateTimeFormat(DISPLAY_LOCALE, { dateStyle: 'medium' })

// A token pasted by hand stands in for Chrome where the app cannot read it, such as in Docker.
export function PastedLogins(): ReactElement {
  const queryClient = useQueryClient()
  const logins = usePastedLogins()
  const [host, setHost] = useState('')
  const [token, setToken] = useState('')
  const refresh = (): Promise<void> =>
    queryClient.invalidateQueries({ queryKey: QUERY_KEYS.pastedLogins })
  const save = useMutation({
    mutationFn: () => savePastedLogin(host, token),
    onSuccess: async () => {
      setToken('')
      await refresh()
    },
  })
  const forget = useMutation({ mutationFn: forgetPastedLogin, onSuccess: refresh })
  const canSave = host.trim() !== '' && token.trim() !== '' && !save.isPending

  return (
    <div className="grid gap-3">
      <h3 className="text-sm font-semibold">Pasted tokens</h3>
      {logins.isError && (
        <p className="text-[13px] text-danger" role="alert">
          {logins.error.message}
        </p>
      )}
      {logins.data !== undefined && logins.data.length > 0 && (
        <ul className="grid list-none gap-1.5 text-[13px]">
          {logins.data.map((saved) => (
            <li key={saved.host} className="flex items-center justify-between gap-3">
              <span className="wrap-anywhere">
                {saved.host} · {saved.userName} · pasted{' '}
                {DATE_FORMAT.format(new Date(saved.savedAt))}
              </span>
              <button
                type="button"
                className="shrink-0 rounded-none bg-transparent p-0 font-medium text-danger"
                disabled={forget.isPending}
                onClick={() => {
                  forget.mutate(saved.host)
                }}
              >
                Delete
              </button>
            </li>
          ))}
        </ul>
      )}
      <form
        className="grid gap-2.5 sm:grid-cols-[1fr_1fr_auto]"
        onSubmit={(event) => {
          event.preventDefault()
          if (canSave) save.mutate()
        }}
      >
        <input
          className={TEXT_FIELD}
          aria-label="Trang Udemy"
          placeholder="tencongty.udemy.com"
          value={host}
          onChange={(event) => {
            setHost(event.target.value)
          }}
        />
        <input
          className={TEXT_FIELD}
          type="password"
          autoComplete="off"
          aria-label="Token access_token"
          placeholder="Paste the access_token value"
          value={token}
          onChange={(event) => {
            setToken(event.target.value)
          }}
        />
        <Button
          type="submit"
          variant={ButtonVariant.Ghost}
          size={ButtonSize.Large}
          icon={Save}
          disabled={!canSave}
        >
          {save.isPending ? 'Checking…' : 'Save token'}
        </Button>
      </form>
      {save.isError && (
        <p className="text-[13px] text-danger" role="alert">
          {save.error.message}
        </p>
      )}
      {forget.isError && (
        <p className="text-[13px] text-danger" role="alert">
          {forget.error.message}
        </p>
      )}
      <p className="text-xs text-muted">
        In Chrome on Udemy: F12 → Application → Cookies → copy access_token. Lasts about 30 days.
      </p>
    </div>
  )
}
