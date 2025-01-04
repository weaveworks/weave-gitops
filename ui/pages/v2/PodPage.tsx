import * as React from "react";
import styled from "styled-components";
import PodDetail from "../../components/PodDetail";
import RequestStateHandler from "../../components/RequestStateHandler";
import { useGetObject } from "../../hooks/objects";
import { Kind } from "../../lib/api/core/types.pb";
import { FluxObject, Pod } from "../../lib/objects";

type Props = {
  className?: string;
  object: FluxObject;
};

function PodPage({ className, object }: Props) {
  const { data, isLoading, error } = useGetObject<Pod>(
    object.name,
    object.namespace,
    Kind.Pod,
    object.clusterName,
    { refetchInterval: false },
  );
  return (
    <RequestStateHandler
      loading={isLoading}
      error={error}
      className={className}
    >
      {data && <PodDetail pod={data} />}
    </RequestStateHandler>
  );
}

export default styled(PodPage).attrs({ className: PodPage.name })`
  width: 100%;
`;
