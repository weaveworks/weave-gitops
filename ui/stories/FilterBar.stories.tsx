import { Meta, Story } from "@storybook/react";
import React from "react";
import FilterBar, { Props } from "../components/FilterBar";

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
  const [storyFilters, setStoryFilters] = React.useState(args.activeFilters);

  return (
    <FilterBar
      {...args}
      activeFilters={storyFilters}
      setActiveFilters={setStoryFilters}
    />
  );
};

export const Default = Template.bind({});
Default.args = {
  className: "",
  activeFilters: ["Ready", "Helm Release"],
  filterList: { Name: ["app", "app2", "app3"], Status: ["Ready", "Failed"] },
};
