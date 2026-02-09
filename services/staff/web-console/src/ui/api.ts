import axios from "axios";

export const api = axios.create({
  baseURL: "/",
  withCredentials: true,
});

export type MeResponse = {
  user: {
    id: string;
    email: string;
    github_login: string;
    is_platform_admin: boolean;
  };
};

export async function getMe(): Promise<MeResponse> {
  const resp = await api.get("/api/v1/auth/me");
  return resp.data as MeResponse;
}

