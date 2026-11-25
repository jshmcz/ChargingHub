import { http } from './request'
import type { Station } from './types'

export interface StationInput {
  tenantId?: number
  name: string
  address: string
  areaCode?: string
  areaName?: string
  defaultStartTime: string
  startThresholdCents: number
  stopThresholdCents: number
  warnThresholdCents: number
  settlementTimeoutHours: number
}

export const stationsApi = {
  list: () => http.get<Station[]>('/stations'),
  create: (body: StationInput) => http.post<Station>('/stations', body),
  update: (id: number, body: StationInput) => http.put<Station>(`/stations/${id}`, body),
}
