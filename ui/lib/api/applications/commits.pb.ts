/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);
export type Commit = {
  commitHash?: string
  date?: string
  author?: string
  message?: string
}


type BaseListCommitsRequest = {
  name?: string
  namespace?: string
  pageSize?: number
}

export type ListCommitsRequest = BaseListCommitsRequest
  & OneOf<{ pageToken: number }>

export type ListCommitsResponse = {
  commits?: Commit[]
  nextPageToken?: number
}