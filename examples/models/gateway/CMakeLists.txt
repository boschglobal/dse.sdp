# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.21)
set(MODEL_NAME "gateway")
project(${MODEL_NAME})

set(EXE_PATH "gateway")

add_executable(gateway
    gateway.c
)
set_target_properties(gateway
    PROPERTIES
        OUTPUT_NAME gateway
        INSTALL_RPATH "$ORIGIN:$ORIGIN/../lib"
)
target_include_directories(${MODEL_NAME}
    PRIVATE
        ${dse_modelc_lib_SOURCE_DIR}/include
)
target_link_directories(${MODEL_NAME}
    PRIVATE
        ${dse_modelc_lib_SOURCE_DIR}/lib
 )
target_compile_definitions(gateway
    PRIVATE
        PLATFORM_OS="${CDEF_PLATFORM_OS}"
        PLATFORM_ARCH="${CDEF_PLATFORM_ARCH}"
        MODELC_VERSION="${VERSION}"
)
target_link_libraries(${MODEL_NAME}
    PRIVATE
        modelc
)
install(
    TARGETS
        gateway
    DESTINATION
        ${EXE_PATH}/bin
    COMPONENT
        gateway_model
)
if(WIN32)
install(
    TARGETS
        ${modelc_shared_lib}
    DESTINATION
        ${EXE_PATH}/bin
    COMPONENT
        gateway_model
)
endif()
install(
    FILES
        simulation.yaml
        model.yaml
        signalgroup.yaml
    DESTINATION
        ${EXE_PATH}/data
    COMPONENT
        gateway_model
)