// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

contract Store {
  event ItemSet(string key);

  mapping (string => bool) public items;

  function setItem(string memory key) external {
    items[key] = true;
    emit ItemSet(key);
  }

  function getItem(string memory key) external view returns (bool) {
    return items[key];
  }
}