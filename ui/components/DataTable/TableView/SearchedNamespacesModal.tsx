import React from "react";
import { SearchedNamespaces } from "../../../lib/types";
import { IconButton } from "../../Button";
import Icon, { IconType } from "../../Icon";
import InfoModal from "../../InfoModal";

interface Props {
  searchedNamespaces: SearchedNamespaces;
}

const SearchedNamespacesModal = ({ searchedNamespaces }: Props) => {
  const [searchedNamespacesModalOpen, setSearchedNamespacesModalOpen] =
    React.useState(false);
  return (
    <>
      <IconButton
        onClick={() =>
          setSearchedNamespacesModalOpen(!searchedNamespacesModalOpen)
        }
        variant="text"
      >
        <Icon type={IconType.InfoIcon} size="medium" color="neutral20" />
      </IconButton>
      <InfoModal
        searchedNamespaces={searchedNamespaces}
        open={searchedNamespacesModalOpen}
        onCloseModal={setSearchedNamespacesModalOpen}
      />
    </>
  );
};

export default SearchedNamespacesModal;
