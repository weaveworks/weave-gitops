import React from "react";
import { useGetObject } from "../../../hooks/objects";
import { Kind } from "../../../lib/api/core/types.pb";
import { FluxObject } from "../../../lib/objects";
import Flex from "../../Flex";
import InfoList from "../../InfoList";
import PageStatus from "../../PageStatus";
import Spacer from "../../Spacer";
import Text from "../../Text";
import LoadingWrapper from "../LoadingWrapper";
type Props = {
  name: string;
  namespace: string;
  clusterName: string;
};
const ImagePolicy = ({ name, namespace, clusterName }: Props) => {
  const { data, isLoading, error } = useGetObject<FluxObject>(
    name,
    namespace,
    Kind.ImagePolicy,
    clusterName,
    {
      refetchInterval: 50000,
    }
  );
  return (
    <LoadingWrapper loading={isLoading} error={error}>
      {!!data && (
        <Flex wide tall column>
          <Text size="large" semiBold titleHeight>
            Policy
          </Text>
          <Spacer margin="xs" />
          <PageStatus conditions={data.conditions} suspended={data.suspended} />
          <Spacer margin="xs" />
          <InfoList
            items={[
              ["Image Policy", Object.keys(data.obj.spec.policy)[0]],
              ["Order/Range", getValueByKey(data.obj.spec.policy, "range")],
              ["Kind", Kind.ImagePolicy],
              ["Name", data.name],
              ["Namespace", data.namespace],
            ]}
          />
        </Flex>
      )}
    </LoadingWrapper>
  );
};

export default ImagePolicy;
function getValueByKey(obj: any, key: string): any {
  const policyKey = Object.keys(obj)[0];
  console.log(policyKey);

  return obj[policyKey][key];
}
