import { Tabs } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import { Container, FluxObject } from "../lib/objects";
import DataTable from "./DataTable";
import Flex from "./Flex";
import InfoList from "./InfoList";
import MuiTab from "./MuiTab";
import PageStatus from "./PageStatus";
import Text from "./Text";
import { DialogYamlView } from "./YamlView";

type Props = {
  className?: string;
  pod: FluxObject;
};

const InfoListUl = styled.ul`
  margin: 0;
  padding: 0;
  list-style: none;
  li {
    padding: 0;
  }
`;

const PodDetailHeader = styled(Text)`
  padding: ${(props) => props.theme.spacing.xs} 0;
  color: ${(props) => props.theme.colors.neutral30};
`;

export const NoDialogDataTable = styled(DataTable)`
  max-width: 100%;
  overflow-x: auto;
`;

const ContainerDivider = styled(Flex)`
  padding-bottom: ${(props) => props.theme.spacing.small};
  margin-bottom: ${(props) => props.theme.spacing.small};
  width: 100%;
  border-bottom: 3px solid;
  border-image-slice: 1;
  border-image-source: ${(props) =>
    `linear-gradient(to right, ${props.theme.colors.neutral30} 0%, ${props.theme.colors.white} 100%)`};
`;

type ListProps = {
  array: any[];
  display: (any) => string;
};
const ArrayToList = ({ array, display }: ListProps) => {
  return (
    <InfoListUl>
      {!array || !array.length ? (
        <li key={0}>-</li>
      ) : (
        array.map((item, i) => <li key={i}>{display(item)}</li>)
      )}
    </InfoListUl>
  );
};
const Detail = ({ pod }) => {
  return (
    <Flex wide column>
      <InfoList
        items={[
          ["Namespace", pod.namespace],
          ["Pod IP", pod.podIP],
          ["Pod IPs", <ArrayToList array={pod.podIPs} display={(p) => p.ip} />],
          ["Priority Class", pod.priorityClass],
          ["QoS Class", pod.qosClass],
        ]}
      />
      <PodDetailHeader bold size="large">
        Tolerations
      </PodDetailHeader>
      <NoDialogDataTable
        hideSearchAndFilters
        fields={[
          { label: "Key", value: "key" },
          { label: "Operator", value: "operator" },
          { label: "Value", value: "value" },
          { label: "Effect", value: "effect" },
          { label: "Seconds", value: "tolerationSeconds" },
        ]}
        rows={pod.tolerations}
      />
      <PodDetailHeader bold size="large">
        Containers
      </PodDetailHeader>
      {pod.containers.map((container: Container, i: number) => {
        return (
          <>
            <InfoList
              key={container.name}
              items={[
                ["Name", container.name],
                ["Image", container.image],
                [
                  "Ports",
                  <ArrayToList
                    array={container.ports}
                    display={(port) =>
                      `${port.name}:${port.containerPort}/${port.protocol}`
                    }
                  />,
                ],
                [
                  "Env Vars",
                  <ArrayToList array={container.enVar} display={(v) => v} />,
                ],
                [
                  "Arguments",
                  <ArrayToList array={container.args} display={(arg) => arg} />,
                ],
              ]}
            />
            {i < pod.containers.length - 1 && <ContainerDivider key={i} />}
          </>
        );
      })}
      <PodDetailHeader bold size="large">
        Volumes
      </PodDetailHeader>
      <ArrayToList
        array={pod.volumes}
        display={({ name, type }) => `${name}: ${type}`}
      />
    </Flex>
  );
};

function PodDetail({ className, pod }: Props) {
  const [tabValue, setTabValue] = React.useState(0);
  const tabKeys = ["detail", "yaml"];

  const tabs = (value: number) => {
    switch (value) {
      case 0:
        return <Detail pod={pod} />;
      case 1:
        return (
          <DialogYamlView
            yaml={pod.yaml}
            object={{
              kind: pod.type,
              name: pod.name,
              namespace: pod.namespace,
              clusterName: pod.clusterName,
            }}
          />
        );
    }
  };
  return (
    <Flex wide tall column className={className}>
      <PageStatus
        conditions={pod.conditions}
        suspended={pod.suspended}
        showAll
      />
      <Tabs
        value={tabValue}
        indicatorColor="primary"
        className="horizontal-tabs"
      >
        {tabKeys.map((key, i) => {
          return (
            <MuiTab
              key={i}
              text={key}
              active={tabValue === i}
              onClick={() => setTabValue(i)}
            />
          );
        })}
      </Tabs>
      {tabs(tabValue)}
    </Flex>
  );
}

export default styled(PodDetail).attrs({ className: PodDetail.name })`
  height: 100%;
  ${PageStatus}, ${NoDialogDataTable} {
    margin-bottom: ${(props) => props.theme.spacing.small};
  }
`;
