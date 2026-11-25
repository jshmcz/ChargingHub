import { get, post } from './request'
import type { Ledger, Wallet } from './types'

export interface WalletMutationPayload {
  userId: string
  amountCents: number
  method: string
  voucherNo?: string
  remark?: string
}

export const walletApi = {
  myWallet(): Promise<Wallet> {
    return get<Wallet>('/my/wallet')
  },

  myLedgers(): Promise<Ledger[]> {
    return get<Ledger[]>('/my/ledgers')
  },

  list(userId?: string): Promise<Wallet[]> {
    return get<Wallet[]>('/wallets', { userId })
  },

  ledgers(userId?: string): Promise<Ledger[]> {
    return get<Ledger[]>('/ledgers', { userId })
  },

  recharge(payload: WalletMutationPayload): Promise<Wallet> {
    return post<Wallet>('/wallet/recharge', payload)
  },

  refund(payload: WalletMutationPayload): Promise<Wallet> {
    return post<Wallet>('/wallet/refund', payload)
  },

  freeze(userId: string, frozen: boolean): Promise<Wallet> {
    return post<Wallet>(`/wallet/${userId}/freeze`, { frozen })
  },
}
