export enum PageRoute {
  Applications = "/applications",
  ApplicationDetail = "/application_detail",
  ApplicationAdd = "/application_add",
  GitlabOAuthCallback = "/oauth/gitlab",
}

export enum GrpcErrorCodes {
  Unknown = 2,
  Unavailable = 14,
  Unauthenticated = 16,
}

export type RequestError = Error & {
  code?: number;
};
