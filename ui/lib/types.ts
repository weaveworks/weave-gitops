export enum PageRoute {
  Applications = "/applications",
  ApplicationDetail = "/application_detail",
}

export enum GrpcErrorCodes {
  Unauthenticated = 16,
}

export enum GitProviderName {
  GitHub = "github",
  Gitlab = "gitlab",
}

export type RequestError = Error & {
  code?: number;
};
