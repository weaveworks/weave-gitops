import { dump } from "js-yaml";
import _ from "lodash";
import {
  Condition,
  GitRepositoryRef,
  GroupVersionKind,
  HealthStatus,
  Interval,
  Kind,
  NamespacedObjectReference,
  ObjectRef,
  Object as ResponseObject,
} from "./api/core/types.pb";

export type Automation = HelmRelease | Kustomization;
export type Source =
  | HelmRepository
  | HelmChart
  | GitRepository
  | Bucket
  | OCIRepository
  | ImageRepository
  | ImageUpdateAutomation;

export interface CrossNamespaceObjectRef extends ObjectRef {
  apiVersion: string;
  matchLabels: { key: string; value: string }[];
}
export interface ImgPolicy {
  type?: string;
  value?: string;
}
export class FluxObject {
  obj: any;
  clusterName: string;
  tenant: string;
  uid: string;
  info: string;
  children: FluxObject[];
  health: HealthStatus;
  isCurrentNode?: boolean;
  constructor(response: ResponseObject | undefined) {
    try {
      this.obj = JSON.parse(response?.payload);
    } catch {
      this.obj = {};
    }
    this.clusterName = response?.clusterName || "";
    this.tenant = response?.tenant || "";
    this.uid = response?.uid || "";
    this.info = response?.info || "";
    this.children = [];
    this.health = response?.health || {};
  }

  get yaml(): string {
    return dump(this.obj);
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

  get labels(): [string, string][] {
    const labels = this.obj.metadata?.labels || {};
    return Object.keys(labels).flatMap((key) => {
      return [[key, labels[key] as string]];
    });
  }

  get suspended(): boolean {
    return Boolean(this.obj.spec?.suspend); // if this is missing, it's not suspended
  }

  get type(): Kind | string | undefined {
    return this.obj.kind || this.obj.groupVersionKind?.kind;
  }

  get conditions(): Condition[] {
    return (
      this.obj.status?.conditions?.map((condition: any) => {
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
        this.obj.spec?.interval,
      );
    const interval = match?.groups || {};
    return {
      hours: interval.hours || "0",
      minutes: interval.minutes || "0",
      seconds: interval.seconds || "0",
    };
  }

  get lastUpdatedAt(): string {
    return this.obj.status?.artifact?.lastUpdateTime || "";
  }

  get images(): string[] {
    const containerPaths = ["spec.template.spec.containers", "spec.containers"];
    const images = containerPaths.flatMap((path) => {
      const containers = _.get(this.obj, path, []);
      // _.map returns an empty list if containers is not iterable
      return _.map(containers, (container: unknown) =>
        _.get(container, "image"),
      );
    });

    // filter out undefined, null, and other strange objects that might be there
    return images.filter((image) => _.isString(image));
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
  get sourceRef(): ObjectRef | undefined {
    if (!this.obj.spec?.sourceRef) {
      return;
    }
    const sourceRef = {
      ...this.obj.spec.sourceRef,
    };
    if (!sourceRef.namespace) {
      sourceRef.namespace = this.namespace;
    }
    return sourceRef;
  }

  get chart(): string {
    return this.obj.spec?.chart || "";
  }

  get version(): string {
    return this.obj.spec?.version || "";
  }

  get revision(): string {
    return this.obj.status?.artifact?.revision || "";
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

  get isVerifiable(): boolean {
    return Boolean(this.obj.spec.verify);
  }
}

export class OCIRepository extends FluxObject {
  get url(): string {
    return this.obj.spec?.url || "";
  }

  get artifactMetadata(): [string, string][] {
    const metadata = this.obj.status?.artifact?.metadata || {};
    const prefix = "org.opencontainers.image/";
    return Object.keys(metadata).flatMap((key) => {
      return [[key.slice(prefix.length), metadata[key] as string]];
    });
  }

  get isVerifiable(): boolean {
    return Boolean(this.obj.spec.verify);
  }
}

export class Kustomization extends FluxObject {
  get dependsOn(): NamespacedObjectReference[] {
    return this.obj.spec?.dependsOn || [];
  }

  get sourceRef(): ObjectRef | undefined {
    if (!this.obj.spec?.sourceRef) {
      return undefined;
    }
    const source = {
      ...this.obj.spec.sourceRef,
    };
    if (!source.namespace) {
      source.namespace = this.namespace;
    }
    return source;
  }

  get path(): string {
    return this.obj.spec?.path || "";
  }

  get lastAppliedRevision(): string {
    return this.obj.status?.lastAppliedRevision || "";
  }

  get inventory(): GroupVersionKind[] {
    const entries = this.obj.status?.inventory?.entries || [];
    return Array.from(
      new Set(
        entries.map((entry) => {
          // entry is namespace_name_group_kind, but name can contain '_' itself
          const parts = entry.id.split("_");
          const kind = parts[parts.length - 1];
          const group = parts[parts.length - 2];
          return { group, version: entry.v, kind };
        }),
      ),
    );
  }
}

export class HelmRelease extends FluxObject {
  inventory: GroupVersionKind[];

  constructor(response: ResponseObject) {
    super(response);
    try {
      this.inventory = response.inventory || [];
    } catch {
      this.inventory = [];
    }
  }

  get dependsOn(): NamespacedObjectReference[] {
    return this.obj.spec?.dependsOn || [];
  }

  get helmChartName(): string {
    return this.obj.status?.helmChart || "";
  }

  get helmChart(): HelmChart {
    // This isn't a "real" helmchart object - it has much fewer fields,
    // and requires some data mangling to work at all
    let chart = this.obj.spec?.chart;
    chart = { ...chart };
    chart.metadata = {
      name: this.namespace + "-" + this.name,
      namespace: chart.spec?.sourceRef?.namespace || this.namespace,
    };
    return new HelmChart({
      payload: JSON.stringify(chart),
      clusterName: this.clusterName,
    });
  }

  get sourceRef(): ObjectRef | undefined {
    return this.helmChart?.sourceRef;
  }

  get lastAppliedRevision(): string {
    return this.obj.status?.lastAppliedRevision || "";
  }

  get lastAttemptedRevision(): string {
    return this.obj.status?.lastAttemptedRevision || "";
  }

  get kubeConfig(): string {
    return this.obj.spec?.kubeConfig?.secretRef?.name || "";
  }
}

export class Provider extends FluxObject {
  get provider(): string {
    return this.obj.spec?.type || "";
  }
  get channel(): string {
    return this.obj.spec?.channel || "";
  }
}

export class ImageUpdateAutomation extends FluxObject {
  get sourceRef(): ObjectRef | undefined {
    if (!this.obj.spec?.sourceRef) {
      return;
    }
    const sourceRef = {
      ...this.obj.spec.sourceRef,
    };
    if (!sourceRef.namespace) {
      sourceRef.namespace = this.namespace;
    }
    return sourceRef;
  }
  get lastAutomationRunTime(): string {
    return this.obj?.status?.lastAutomationRunTime || "";
  }
}
export class ImagePolicy extends ImageUpdateAutomation {
  constructor(response: ResponseObject) {
    super(response);
  }
  get imagePolicy(): ImgPolicy {
    const { policy } = this.obj?.spec || {};
    const [type] = Object.keys(policy);
    if (type) {
      const [val] = Object.values(policy[type]);
      return {
        type,
        value: (val as string) || "",
      };
    }
    return {
      type: "",
      value: "",
    };
  }
  get imageRepositoryRef(): string {
    return this.obj?.spec?.imageRepositoryRef?.name || "";
  }
}
export class ImageRepository extends ImageUpdateAutomation {
  constructor(response: ResponseObject) {
    super(response);
  }
  get tagCount(): string {
    return this.obj.status?.lastScanResult?.tagCount || "";
  }
}
export class Alert extends FluxObject {
  get providerRef(): string {
    return this.obj.spec?.providerRef.name || "";
  }
  get severity(): string {
    return this.obj.spec?.eventSeverity || "";
  }
  get eventSources(): CrossNamespaceObjectRef[] {
    return this.obj.spec?.eventSources || [];
  }
}

//for pods
export type Toleration = {
  key: string;
  operator: string;
  value: string;
  effect: string;
  tolerationSeconds: number;
};

export type Container = {
  name: string;
  image: string;
  args: string[];
  ports: any[];
  enVar: string[];
  //status?
};

export class Pod extends FluxObject {
  get podIP(): string {
    return this.obj.status?.podIP || "-";
  }
  get podIPs(): string[] {
    return this.obj.status?.podIPs || ["-"];
  }
  get priorityClass(): string {
    return this.obj.spec?.priorityClassName || "-";
  }
  get qosClass(): string {
    return this.obj.status?.qosClass || "-";
  }
  get tolerations(): Toleration[] {
    return this.obj.spec?.tolerations || [];
  }
  get containers(): Container[] {
    return this.obj.spec?.containers || [];
  }
  get volumes(): { name: string; type: string }[] {
    const volumeObjs = [];
    const volumes = this.obj.spec?.volumes || [];
    volumes.forEach((volume) => {
      const name = volume.name || "-";
      let type = "-";
      Object.keys(volume).forEach((key) => {
        if (key !== "name" && key !== "emptyDir") type = key;
      });
      volumeObjs.push({ name, type });
    });
    return volumeObjs;
  }
}

export function makeObjectId(namespace?: string, name?: string) {
  return namespace + "/" + name;
}

export type FluxObjectNodesMap = { [key: string]: FluxObjectNode };

export class FluxObjectNode {
  obj: any;
  uid: string;
  type?: string;
  name: string;
  namespace: string;
  clusterName: string;
  suspended: boolean;
  conditions: Condition[];
  dependsOn: NamespacedObjectReference[];
  isCurrentNode?: boolean;
  yaml: string;
  id: string;
  parentIds: string[];

  constructor(fluxObject: FluxObject, isCurrentNode?: boolean) {
    this.obj = fluxObject.obj;
    this.uid = fluxObject.uid;
    this.type = fluxObject.type;
    this.name = fluxObject.name;
    this.namespace = fluxObject.namespace;
    this.clusterName = fluxObject.clusterName;
    this.suspended = fluxObject.suspended;
    this.conditions = fluxObject.conditions;
    this.dependsOn =
      (fluxObject as Kustomization | HelmRelease).dependsOn || [];
    this.isCurrentNode = isCurrentNode;
    this.yaml = fluxObject.yaml;
    this.id = makeObjectId(this.namespace, this.name);
    this.parentIds = this.dependsOn.map((dependency) => {
      const namespace = dependency.namespace || this.namespace;

      return namespace + "/" + dependency.name;
    });
  }
}
