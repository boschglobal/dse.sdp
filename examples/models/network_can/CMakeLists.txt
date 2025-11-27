# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.21)
cmake_policy(SET CMP0135 NEW)
project(network)

set(NETWORK_CAN "network_can")

add_library(${NETWORK_CAN} SHARED
    model_can.c
)
target_include_directories(${NETWORK_CAN}
    PRIVATE
        ${dse_modelc_lib_SOURCE_DIR}/include
)
target_link_directories(${NETWORK_CAN}
    PRIVATE
        ${dse_modelc_lib_SOURCE_DIR}/lib
 )
target_link_libraries(${NETWORK_CAN}
    PRIVATE
        modelc
)
install(TARGETS ${NETWORK_CAN}
    LIBRARY DESTINATION
        ${NETWORK_CAN}/lib
    COMPONENT
        ${NETWORK_CAN}
)
install(
    FILES
        model.yaml
        simulation.yaml
    DESTINATION
        ${NETWORK_CAN}/data
    COMPONENT
        ${NETWORK_CAN}
)
