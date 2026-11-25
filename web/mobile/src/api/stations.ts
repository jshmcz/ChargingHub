import { get, post, put } from './request'
import type { Gun, Pile, PilePayload, Station, StationPayload } from './types'

export const stationApi = {
  /** 用户端 / 管理端共用，返回本租户场站（含枪状态） */
  list(): Promise<Station[]> {
    return get<Station[]>('/stations')
  },

  guns(stationId: string): Promise<Gun[]> {
    return get<Gun[]>('/guns', { stationId })
  },

  create(payload: StationPayload): Promise<Station> {
    return post<Station>('/stations', payload)
  },

  update(id: string, payload: StationPayload): Promise<Station> {
    return put<Station>(`/stations/${id}`, payload)
  },

  piles(): Promise<Pile[]> {
    return get<Pile[]>('/piles')
  },

  createPile(payload: PilePayload): Promise<Pile> {
    return post<Pile>('/piles', payload)
  },

  updatePile(id: string, status: string): Promise<Pile> {
    return put<Pile>(`/piles/${id}`, { status })
  },

  regenerateSecret(id: string): Promise<{ secret: string }> {
    return post<{ secret: string }>(`/piles/${id}/regenerate-secret`)
  },
}
