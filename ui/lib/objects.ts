import { stringify } from "yaml";
import {
  Condition,
  FluxObjectKind,
  FluxObjectRef,
  GitRepositoryRef,
  Interval,
  NamespacedObjectReference,
  Object as ResponseObject,
  ObjectRef,
} from "./api/core/types.pb";
import { addKind } from "./utils";

export enum Kind {
  GitRepository = "GitRepository",
  Bucket = "Bucket",
  HelmRepository = "HelmRepository",
  HelmChart = "HelmChart",
  Kustomization = "Kustomization",
  HelmRelease = "HelmRelease",
  OCIRepository = "OCIRepository",
  Provider = "Provider",
  Alert = "Alert",
}

export type Source =
  | HelmRepository
  | HelmChart
  | GitRepository
  | Bucket
  | OCIRepository;

export function fluxObjectKindToKind(fok: FluxObjectKind): Kind {
  return Kind[FluxObjectKind[fok].slice(4)];
}

export interface CrossNamespaceObjectRef extends ObjectRef {
  apiVersion: string;
  matchLabels: { key: string; value: string }[];
}

export class FluxObject {
  obj: any;
  clusterName: string;
  tenant: string;
  uid: string;

  constructor(response: ResponseObject) {
    try {
      this.obj = JSON.parse(response.payload);
    } catch {
      this.obj = {};
    }
    this.clusterName = response?.clusterName;
    this.tenant = response?.tenant;
    this.uid = response?.uid;
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

  // TODO: this actually returns the k8s kind name,
  // while kind returns a value with a non-standard name.
  // We shouldn't need both, and this value should be sufficient
  get type(): Kind | undefined {
    return this.obj.kind;
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
  get sourceRef(): FluxObjectRef | undefined {
    if (!this.obj.spec?.sourceRef) {
      return;
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

export class OCIRepository extends FluxObject {
  get url(): string {
    return this.obj.spec?.url || "";
  }

  get source(): string {
    const metadata = this.obj.status?.artifact?.metadata;
    if (!metadata) {
      return "";
    }
    return metadata["org.opencontainers.image.source"] || "";
  }

  get revision(): string {
    const metadata = this.obj.status?.artifact?.metadata;
    if (!metadata) {
      return "";
    }
    return metadata["org.opencontainers.image.revision"] || "";
  }
}

export class Kustomization extends FluxObject {
  get dependsOn(): NamespacedObjectReference[] {
    return this.obj.spec?.dependsOn || [];
  }
}

export class HelmRelease extends FluxObject {
  get dependsOn(): NamespacedObjectReference[] {
    return this.obj.spec?.dependsOn || [];
  }
}

export class Provider extends FluxObject {
  get provider(): string {
    return this.obj.spec.type || "";
  }
  get channel(): string {
    return this.obj.spec.channel || "";
  }
}

export class Alert extends FluxObject {
  get providerRef(): string {
    return this.obj.spec.providerRef.name || "";
  }
  get severity(): string {
    return this.obj.spec.eventSeverity || "";
  }
  get eventSources(): CrossNamespaceObjectRef[] {
    return this.obj.spec.eventSources || [];
  }
}

export function makeObjectId(namespace?: string, name?: string) {
  return namespace + "/" + name;
}

export class FluxObjectNode {
  obj: any;
  uid: string;
  displayKind?: string;
  name: string;
  namespace: string;
  suspended: boolean;
  conditions: Condition[];
  dependsOn: NamespacedObjectReference[];
  isCurrentNode?: boolean;
  id: string;
  parentIds: string[];

  constructor(fluxObject: FluxObject, isCurrentNode?: boolean) {
    this.obj = fluxObject.obj;
    this.uid = fluxObject.uid;
    this.displayKind = fluxObject.type;
    this.name = fluxObject.name;
    this.namespace = fluxObject.namespace;
    this.suspended = fluxObject.suspended;
    this.conditions = fluxObject.conditions;
    this.dependsOn =
      (fluxObject as Kustomization | HelmRelease).dependsOn || [];
    this.isCurrentNode = isCurrentNode;
    this.id = makeObjectId(this.namespace, this.name);
    this.parentIds = this.dependsOn.map((dependency) => {
      const namespace = dependency.namespace || this.namespace;

      return namespace + "/" + dependency.name;
    });
  }
}
