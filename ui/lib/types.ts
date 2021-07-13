export enum PageRoute {
  Auth = "/auth",
  OAuthCallback = "/oauth_callback",
  Applications = "/applications",
  ApplicationDetail = "/application_detail",
}

export type AsyncError = {
  message: string;
  detail: string;
};
