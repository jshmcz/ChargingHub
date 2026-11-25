import { http } from './request'
import type { Gun, Pile, PileStatus } from './types'

export interface PileInput {
  stationId: number
  deviceNo: string
  gunCount: number
}

export const pilesApi = {
  list: () => http.get<Pile[]>('/piles'),
  create: (body: PileInput) => http.post<Pile>('/piles', body),
  updateStatus: (id: number, status: PileStatus) => http.put<Pile>(`/piles/${id}`, { status }),
  regenerateSecret: (id: number) => http.post<{ secret: string }>(`/piles/${id}/regenerate-secret`),
}

export const gunsApi = {
  list: (stationId: number) => http.get<Gun[]>('/guns', { stationId }),
}
