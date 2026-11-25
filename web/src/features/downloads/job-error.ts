import type { Job } from '../../api/types'

type JobErrorKind = NonNullable<Job['errorKind']>

const ERROR_ADVICE: Readonly<Record<Exclude<JobErrorKind, 'other'>, string>> = {
  login: 'Đăng nhập Udemy đã hết hạn. Dán token mới rồi tải lại.',
  course_not_found:
    'Không thấy khoá này trong tài khoản Udemy. Kiểm tra link, hoặc mở khoá một lần trên trình duyệt rồi tải lại.',
  blocked:
    'Udemy đang chặn yêu cầu từ app (Cloudflare). Thử lại sau; nếu vẫn bị, cần cập nhật udemy-studio.',
  chrome_access:
    'Không đọc được đăng nhập trong Chrome. Cấp quyền cho app hoặc dán token rồi tải lại.',
  interrupted: 'Server đã tắt giữa chừng. Tải lại để làm tiếp phần còn thiếu.',
}

export function jobErrorText(job: Job): string | undefined {
  if (job.error === undefined) return undefined
  if (job.errorKind === undefined || job.errorKind === 'other') return job.error
  return ERROR_ADVICE[job.errorKind]
}
