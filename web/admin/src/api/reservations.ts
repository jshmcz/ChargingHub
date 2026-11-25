import { http } from './request'
import type { Reservation } from './types'

export const reservationsApi = {
  list: (userId?: number) => http.get<Reservation[]>('/reservations', { userId }),
  cancel: (id: number) => http.post<null>(`/reservations/${id}/cancel`),
  my: () => http.get<Reservation[]>('/my/reservations'),
}
