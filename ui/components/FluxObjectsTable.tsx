import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
import { useLinkResolver } from "../contexts/LinkResolverContext";
import { Kind } from "../lib/api/core/types.pb";
import { formatURL, objectTypeToRoute } from "../lib/nav";
import { FluxObject } from "../lib/objects";
import { makeImageString, statusSortHelper } from "../lib/utils";
import DataTable from "./DataTable";
import { DetailViewProps } from "./DetailModal";
import HealthCheckStatusIndicator, {
  HealthStatusType,
} from "./HealthCheckStatusIndicator";
import ImageLink from "./ImageLink";
import KubeStatusIndicator, {
  computeMessage,
  createSyntheticCondition,
  ReadyStatusValue,
  SpecialObject,
} from "./KubeStatusIndicator";
import Link from "./Link";
import Text from "./Text";

type Props = {
  className?: string;
  onClick?: (o: DetailViewProps) => void;
  objects: FluxObject[];
  initialFilterState?: any;
};

function FluxObjectsTable({
  className,
  onClick,
  initialFilterState,
  objects,
}: Props) {
  const resolver = useLinkResolver();

  return (
    <DataTable
      filters={initialFilterState}
      className={className}
      fields={[
        {
          value: (u: FluxObject) => {
            const kind = Kind[u.type];
            const secret = u.type === "Secret";
            const params = {
              name: u.name,
              namespace: u.namespace,
              clusterName: u.clusterName,
            };
            // Enterprise is "aware" of more types of objects than Core,
            // and we want to be able to link to those within this table.
            // The resolver func provided by the context will decide what URL this routes to.
            const resolved = resolver && resolver(u.type, params);
            const route = objectTypeToRoute(kind);
            const formatted = formatURL(route, params);

            if (route || resolved) {
              return <Link to={resolved || formatted}>{u.name}</Link>;
            }

            return (
              <Text
                onClick={() =>
                  secret
                    ? null
                    : onClick({
                        object: u,
                      })
                }
                color={secret ? "neutral40" : "primary10"}
                pointer={!secret}
              >
                {u.name}
              </Text>
            );
          },
          label: "Name",
          sortValue: (u: FluxObject) => u.name || "",
          textSearchable: true,
          maxWidth: 600,
        },
        {
          label: "Kind",
          value: (u: FluxObject) => u.type,
          sortValue: (u: FluxObject) => u.type,
        },
        {
          label: "Namespace",
          value: "namespace",
          sortValue: ({ namespace }) => namespace,
        },
        {
          label: "Health Check",
          value: ({ health }) => {
            return health.status !== HealthStatusType.Unknown ? (
              <HealthCheckStatusIndicator health={health} />
            ) : null;
          },
          sortValue: ({ health }: FluxObject) => health.status,
        },
        {
          label: "Status",
          value: (u: FluxObject) => {
            const status = u.obj.status;

            if (!status || !status.conditions) {
              const cond = createSyntheticCondition(
                u.type as SpecialObject,
                status
              );

              if (cond.status === ReadyStatusValue.Unknown) {
                return null;
              }

              return (
                <KubeStatusIndicator
                  conditions={[cond]}
                  suspended={u.suspended}
                  short
                />
              );
            }

            return u.conditions.length > 0 ? (
              <KubeStatusIndicator
                conditions={u.conditions}
                suspended={u.suspended}
                short
              />
            ) : null;
          },
          sortValue: statusSortHelper,
        },
        {
          label: "Message",
          value: (u: FluxObject) => _.first(u.conditions)?.message,
          sortValue: ({ conditions }) => computeMessage(conditions),
          maxWidth: 600,
        },
        {
          label: "Images",
          value: (u: FluxObject) => (
            <ImageLink image={makeImageString(u.images)} />
          ),
          sortValue: (u: FluxObject) => makeImageString(u.images),
        },
      ]}
      rows={objects}
    />
  );
}
export default styled(FluxObjectsTable).attrs({
  className: FluxObjectsTable.name,
})``;
