import { http } from './request'
import type { Ledger, Wallet } from './types'

export interface WalletOpPayload {
  userId: number
  amountCents: number
  method: string
  voucherNo: string
  remark: string
}

export const walletsApi = {
  list: (userId?: number) => http.get<Wallet[]>('/wallets', { userId }),
  recharge: (body: WalletOpPayload) => http.post<Wallet>('/wallet/recharge', body),
  refund: (body: WalletOpPayload) => http.post<Wallet>('/wallet/refund', body),
  freeze: (userId: number, frozen: boolean) => http.post<Wallet>(`/wallet/${userId}/freeze`, { frozen }),
  my: () => http.get<Wallet>('/my/wallet'),
}

export const ledgersApi = {
  list: (userId?: number) => http.get<Ledger[]>('/ledgers', { userId }),
  my: () => http.get<Ledger[]>('/my/ledgers'),
}
