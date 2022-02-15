import { Meta, Story } from "@storybook/react";
import React from "react";
import FilterBar, { Props } from "../components/FilterBar";
import Flex from "../components/Flex";

export default {
  title: "FilterBar",
  component: FilterBar,
  parameters: {
    docs: {
      description: {
        component: "Filters row data to be passed down to `<DataTable />`",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => {
  const [storyFilters, setStoryFilters] = React.useState([]);

  return (
    <Flex wide align end>
      <FilterBar
        {...args}
        activeFilters={storyFilters}
        setActiveFilters={setStoryFilters}
      />
    </Flex>
  );
};

export const Default = Template.bind({});
Default.args = {
  filterList: {
    Name: ["app", "app2", "app3"],
    Status: ["Ready", "Failed"],
    Type: ["Application", "Helm Release"],
  },
};
