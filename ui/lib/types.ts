export enum PageRoute {
  Applications = "/applications",
  ApplicationDetail = "/application_detail",
  ApplicationAdd = "/application_add",
  ApplicationRemove = "/application_remove",
  GitlabOAuthCallback = "/oauth/gitlab",
}

export enum GrpcErrorCodes {
  Unauthenticated = 16,
}

export type RequestError = Error & {
  code?: number;
};
