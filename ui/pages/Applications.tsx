import { Checkbox } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Button from "../components/Button";
import DataTable from "../components/DataTable";
import Flex from "../components/Flex";
import Icon, { IconType } from "../components/Icon";
import Link from "../components/Link";
import Page, { TitleBar } from "../components/Page";
import useApplications from "../hooks/applications";
import { Application } from "../lib/api/applications/applications.pb";
import { PageRoute } from "../lib/types";
import { formatURL } from "../lib/utils";
type Props = {
  className?: string;
};

function Applications({ className }: Props) {
  const [applications, setApplications] = React.useState<Application[]>([]);
  const [sort, setSort] = React.useState("Name");
  const [reverseSort, setReverseSort] = React.useState(false);
  const { listApplications, loading } = useApplications();

  React.useEffect(() => {
    listApplications().then((res) => setApplications(res as Application[]));
  }, []);

  type labelProps = { label: string };
  function SortableLabel({ label }: labelProps) {
    return (
      <Flex align>
        <Button
          onClick={() => {
            setSort(label);
            setReverseSort(false);
          }}
          className={`lowercase table ${sort === label && "bold"}`}
        >
          <p>{label}</p>
        </Button>
        {sort === label && (
          <Button
            onClick={() =>
              reverseSort ? setReverseSort(false) : setReverseSort(true)
            }
            className="table"
          >
            <Icon
              type={IconType.ArrowDownward}
              size="small"
              className={reverseSort ? "upward" : "downward"}
            />
          </Button>
        )}
      </Flex>
    );
  }

  return (
    <Page loading={loading} className={className}>
      <Flex align between>
        <TitleBar>
          <h2>Installed GitOps Applications</h2>
        </TitleBar>
      </Flex>
      <Flex className="application-actions">
        <Link to={PageRoute.ApplicationAdd}>
          <Button
            variant="outlined"
            color="primary"
            type="button"
            className="applications bold"
          >
            Add a new app <Icon type={IconType.Add} size="base" />
          </Button>
        </Link>
        <Button
          variant="outlined"
          color="primary"
          type="button"
          className="applications bold"
        >
          Install a new Profile <Icon type={IconType.Add} size="base" />
        </Button>
        <Button
          variant="outlined"
          color="secondary"
          type="button"
          className="applications bold"
        >
          Create a PR to Delete App{" "}
          <Icon type={IconType.DeleteForever} size="base" />
        </Button>
      </Flex>

      <DataTable
        sortFields={[sort]}
        reverseSort={reverseSort}
        fields={[
          {
            label: <Checkbox />,
            value: () => <Checkbox />,
          },
          {
            label: <SortableLabel label="Name" />,
            value: ({ name }: Application) => (
              <Link to={formatURL(PageRoute.ApplicationDetail, { name })}>
                {name}
              </Link>
            ),
          },
          {
            label: "Status",
            value: "status",
          },
          {
            label: <SortableLabel label="Type" />,
            value: "type",
          },
          {
            label: "Version",
            value: "version",
          },
          {
            label: <SortableLabel label="Namespace" />,
            value: ({ namespace }: Application) => (
              <p>{namespace ? namespace : "default"}</p>
            ),
          },
          // Probably going to need this eventually, but we don't have a status
          // for an app from the backend yet. Keep the code around to avoid
          // having to figure this out again.
          // {
          //   label: "Status",
          //   value: () => (
          //     <Icon
          //       size="medium"
          //       color="success"
          //       type={IconType.CheckMark}
          //       text="Ready"
          //     />
          //   ),
          // },
        ]}
        rows={applications}
      />
    </Page>
  );
}

export default styled(Applications)``;
