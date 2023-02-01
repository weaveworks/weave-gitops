import _ from "lodash";
import React from "react";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { FluxObject } from "../../../lib/objects";
import Flex from "../../Flex";
import InfoList, { InfoField } from "../../InfoList";
import PageStatus from "../../PageStatus";
import RequestStateHandler from "../../RequestStateHandler";
import Spacer from "../../Spacer";
import Text from "../../Text";

type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};

function getInfoItems(data: FluxObject): InfoField[] {
  const [imgPolicy] = Object.keys(_.get(data, ["obj", "spec", "policy"]));
  const val = Object.values(_.get(data, ["obj", "spec", "policy", imgPolicy]));

  return [
    ["Image Policy", imgPolicy],
    ["Order/Range", val],
    ["Kind", Kind.ImagePolicy],
    ["Name", data.name],
    ["Namespace", data.namespace],
  ];
}

const ImagePolicy = ({ name, namespace, clusterName }: Props) => {
  const { data, isLoading, error } = useGetObject<FluxObject>(
    name,
    namespace,
    Kind.ImagePolicy,
    clusterName,
    {
      refetchInterval: 5000,
    }
  );
  return (
    <RequestStateHandler loading={isLoading} error={error}>
      {!!data && (
        <Flex wide tall column>
          <Text size="large" semiBold titleHeight>
            Policy
          </Text>
          <Spacer margin="xs" />
          <PageStatus conditions={data.conditions} suspended={data.suspended} />
          <Spacer margin="xs" />
          <InfoList items={getInfoItems(data)} />
        </Flex>
      )}
    </RequestStateHandler>
  );
};

export default ImagePolicy;
