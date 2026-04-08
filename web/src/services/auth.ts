import { apiRequest } from '@/services/api';
import type { LoginPayload, LoginResponse } from '@/types/auth';

export function login(payload: LoginPayload) {
  return apiRequest<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

