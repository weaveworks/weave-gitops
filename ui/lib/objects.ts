import { stringify } from "yaml";
import { FluxObjectKind, Object as ResponseObject } from "./api/core/types.pb";

export enum Kind {
  GitRepository = "GitRepository",
  Bucket = "Bucket",
  HelmRepository = "HelmRepository",
  HelmChart = "HelmChart",
  Kustomization = "Kustomization",
  HelmRelease = "HelmRelease",
}

export function fluxObjectKindToKind(fok: FluxObjectKind): Kind {
  return Kind[FluxObjectKind[fok].slice(4)];
}

export class FluxObject {
  obj: any;

  constructor(response: ResponseObject) {
    this.obj = JSON.parse(response.payload);
  }

  yaml() {
    return stringify(this.obj);
  }

  // Return list of key-value pairs for the metadata annotations that follow
  // our spec
  metadata(): [string, string][] {
    const prefix = "metadata.weave.works/";
    const annotations = this.obj.metadata.annotations || {};
    return Object.keys(annotations).flatMap((key) => {
      if (!key.startsWith(prefix)) {
        return [];
      } else {
        return [[key.slice(prefix.length), annotations[key] as string]];
      }
    });
  }
}
