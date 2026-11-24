import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { forgetPastedLogin, savePastedLogin, usePastedLogins } from '../../api/download-queries'
import { QUERY_KEYS } from '../../api/queries'

const DATE_FORMAT = new Intl.DateTimeFormat('vi-VN', { dateStyle: 'medium' })
const FIELD = 'rounded-lg border border-line bg-page px-3 py-2'

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
      <h3 className="m-0 text-[0.95rem] font-semibold">Token dán tay</h3>
      {logins.isError && <p className="m-0 text-danger">{logins.error.message}</p>}
      {logins.data !== undefined && logins.data.length > 0 && (
        <ul className="m-0 grid list-none gap-1 p-0 text-[0.9rem]">
          {logins.data.map((saved) => (
            <li key={saved.host} className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <span>
                {saved.host} · {saved.userName}
                <span className="text-muted">
                  {' '}
                  · dán ngày {DATE_FORMAT.format(new Date(saved.savedAt))}
                </span>
              </span>
              <button
                type="button"
                className="link-button text-danger"
                disabled={forget.isPending}
                onClick={() => {
                  forget.mutate(saved.host)
                }}
              >
                Xoá
              </button>
            </li>
          ))}
        </ul>
      )}
      <form
        className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]"
        onSubmit={(event) => {
          event.preventDefault()
          if (canSave) save.mutate()
        }}
      >
        <input
          className={FIELD}
          aria-label="Trang Udemy"
          placeholder="tencongty.udemy.com"
          value={host}
          onChange={(event) => {
            setHost(event.target.value)
          }}
        />
        <input
          className={FIELD}
          type="password"
          autoComplete="off"
          aria-label="Token access_token"
          placeholder="Dán giá trị access_token"
          value={token}
          onChange={(event) => {
            setToken(event.target.value)
          }}
        />
        <button type="submit" disabled={!canSave}>
          {save.isPending ? 'Đang kiểm tra…' : 'Lưu token'}
        </button>
      </form>
      {save.isError && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {save.error.message}
        </p>
      )}
      {forget.isError && (
        <p className="m-0 text-[0.9rem] text-danger" role="alert">
          {forget.error.message}
        </p>
      )}
      <p className="m-0 text-[0.85rem] text-muted">
        Lấy trong Chrome: mở trang Udemy, bấm F12, vào Application → Cookies, chép giá trị dòng
        access_token. Token dùng được khoảng 30 ngày; khi hết hạn, lượt tải sẽ báo lỗi đăng nhập và
        chỉ cần dán token mới.
      </p>
    </div>
  )
}
