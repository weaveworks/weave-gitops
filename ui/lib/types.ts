export enum PageRoute {
  Applications = "/applications",
  ApplicationDetail = "/application_detail",
  ApplicationAdd = "/application_add",
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
