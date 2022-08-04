import { stringify } from "yaml";
import { addKind } from "./utils";
import {
  GitRepositoryRef,
  Condition,
  FluxObjectKind,
  FluxObjectRef,
  Object as ResponseObject,
  Interval,
} from "./api/core/types.pb";

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
  clusterName: string;

  constructor(response: ResponseObject) {
    try {
      this.obj = JSON.parse(response.payload);
    } catch {
      this.obj = {};
    }
    this.clusterName = response.clusterName;
  }

  get yaml(): string {
    return stringify(this.obj);
  }

  get name(): string {
    return this.obj.metadata?.name || "";
  }

  get namespace(): string {
    return this.obj.metadata?.namespace || "";
  }

  // Return list of key-value pairs for the metadata annotations that follow
  // our spec
  get metadata(): [string, string][] {
    const prefix = "metadata.weave.works/";
    const annotations = this.obj.metadata?.annotations || {};
    return Object.keys(annotations).flatMap((key) => {
      if (!key.startsWith(prefix)) {
        return [];
      } else {
        return [[key.slice(prefix.length), annotations[key] as string]];
      }
    });
  }

  get suspended(): boolean {
    return Boolean(this.obj.spec?.suspend); // if this is missing, it's not suspended
  }

  get kind(): FluxObjectKind | undefined {
    if (this.obj.kind) {
      return FluxObjectKind[addKind(this.obj.kind)];
    }
  }

  get conditions(): Condition[] {
    return (
      this.obj.status?.conditions?.map((condition) => {
        return {
          type: condition.type,
          status: condition.status,
          reason: condition.reason,
          message: condition.message,
          timestamp: condition.lastTransitionTime,
        };
      }) || []
    );
  }

  get interval(): Interval {
    const match =
      /((?<hours>[0-9]+)h)?((?<minutes>[0-9]+)m)?((?<seconds>[0-9]+)s)?/.exec(
        this.obj.spec?.interval
      );
    const interval = match.groups;
    return {
      hours: interval.hours || "0",
      minutes: interval.minutes || "0",
      seconds: interval.seconds || "0",
    };
  }

  get lastUpdatedAt(): string {
    return this.obj.status?.artifact?.lastUpdateTime || "";
  }
}

export class HelmRepository extends FluxObject {
  get repositoryType(): string {
    return this.obj.spec?.type == "oci" ? "OCI" : "Default";
  }

  get url(): string {
    return this.obj.spec?.url || "";
  }
}

export class HelmChart extends FluxObject {
  get sourceRef(): FluxObjectRef {
    if (!this.obj.spec?.sourceRef) {
      return {};
    }
    const sourceRef = {
      ...this.obj.spec.sourceRef,
      kind: FluxObjectKind[addKind(this.obj.spec.sourceRef.kind)],
    };
    if (!sourceRef.namespace) {
      sourceRef.namespace = this.namespace;
    }
    return sourceRef;
  }

  get chart(): string {
    return this.obj.spec?.chart || "";
  }
}

export class Bucket extends FluxObject {
  get endpoint(): string {
    return this.obj.spec?.endpoint || "";
  }
}

export class GitRepository extends FluxObject {
  get url(): string {
    return this.obj.spec?.url || "";
  }

  get reference(): GitRepositoryRef {
    return this.obj.spec?.ref || {};
  }
}
