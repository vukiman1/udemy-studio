import { useMutation } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import { CircleCheck, LoaderCircle } from 'lucide-react'
import { checkStorage } from '../../api/storage-queries'
import type { StorageAccess, StorageCheck } from '../../api/types'
import { Button } from '../../ui/Button'
import { SETTINGS_ROW } from '../../ui/settings-row'
import type { StorageDraftEditor } from './use-storage-draft'

const CHECKING_ICON_CLASS = '[&_svg]:animate-spin'

const ACCESS_TEXT: Readonly<Record<StorageAccess, string>> = {
  'read-write': 'Đọc–ghi: bản trên Mac đẩy khoá học lên bucket này được.',
  'read-only': 'Chỉ đọc: hợp với bản học trên VPS, không đẩy khoá lên được.',
}

interface StorageCheckFieldProps {
  readonly editor: StorageDraftEditor
}

export function StorageCheckField({ editor }: StorageCheckFieldProps): ReactElement {
  const check = useMutation({ mutationFn: checkStorage })
  const isForCurrentDraft = check.variables === editor.draft

  return (
    <div className={SETTINGS_ROW}>
      <span aria-hidden="true" />
      <div className="grid justify-items-start gap-3">
        <Button
          icon={check.isPending ? LoaderCircle : CircleCheck}
          className={check.isPending ? CHECKING_ICON_CLASS : ''}
          disabled={check.isPending || !editor.isComplete}
          onClick={() => {
            check.mutate(editor.draft)
          }}
        >
          {check.isPending ? 'Đang kiểm tra…' : 'Kiểm tra kết nối'}
        </Button>
        {isForCurrentDraft && check.isSuccess && <CheckResult result={check.data} />}
        {isForCurrentDraft && check.isError && (
          <p className="text-[13px] text-danger" role="alert">
            {check.error.message}
          </p>
        )}
      </div>
    </div>
  )
}

interface CheckResultProps {
  readonly result: StorageCheck
}

function CheckResult({ result }: CheckResultProps): ReactElement {
  return (
    <p className="text-[13px] text-muted">
      <span className="text-accent-soft">● Kết nối được.</span>
      {` ${ACCESS_TEXT[result.access]} `}
      {result.hasCourse ? 'Bucket đang có một khoá học.' : 'Bucket chưa có khoá học nào.'}
    </p>
  )
}
