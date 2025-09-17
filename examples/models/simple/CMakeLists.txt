# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.21)
set(MODEL_NAME "simple")
project(${MODEL_NAME})

add_library(${MODEL_NAME} SHARED
    model.c
)
target_include_directories(${MODEL_NAME}
    PRIVATE
        ${dse_modelc_lib_SOURCE_DIR}/include
)
target_link_directories(${MODEL_NAME}
    PRIVATE
        ${dse_modelc_lib_SOURCE_DIR}/lib
 )
target_link_libraries(${MODEL_NAME}
    PRIVATE
        modelc
)
install(TARGETS ${MODEL_NAME}
    LIBRARY DESTINATION
        ${MODEL_NAME}/lib
    COMPONENT
        ${MODEL_NAME}
)
install(
    FILES
        model.yaml
        simulation.yaml
    DESTINATION
        ${MODEL_NAME}/data
    COMPONENT
        ${MODEL_NAME}
)
