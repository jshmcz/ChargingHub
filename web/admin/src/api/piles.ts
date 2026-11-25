import { http } from './request'
import type { Gun, Pile, PileStatus } from './types'

export interface PileInput {
  stationId: number
  gunCount: number
}

export interface PileResult extends Pile {
  gunCodes?: string[]
  qrTokens?: string[]
}

export const pilesApi = {
  list: () => http.get<Pile[]>('/piles'),
  create: (body: PileInput) => http.post<PileResult>('/piles', body),
  updateStatus: (id: number, status: PileStatus) => http.put<Pile>(`/piles/${id}`, { status }),
  regenerateSecret: (id: number) => http.post<{ secret: string }>(`/piles/${id}/regenerate-secret`),
  downloadQrcode: (gunId: number, filename?: string) =>
    http.download(`/guns/${gunId}/qrcode`, filename || `gun-${gunId}.png`),
}

export const gunsApi = {
  list: (stationId: number) => http.get<Gun[]>('/guns', { stationId }),
}
