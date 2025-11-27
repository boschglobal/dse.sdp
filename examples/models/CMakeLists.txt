# Copyright 2025 Robert Bosch GmbH
#
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.21)
cmake_policy(SET CMP0135 NEW)

if ($ENV{PACKAGE_ARCH} MATCHES "linux-amd64")
    set(CMAKE_SYSTEM_NAME Linux)
elseif ($ENV{PACKAGE_ARCH} MATCHES "linux-i386")
    set(CMAKE_SYSTEM_NAME Linux)
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -m32")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -m32")
elseif ($ENV{PACKAGE_ARCH} MATCHES "linux-x86")
    set(CMAKE_SYSTEM_NAME Linux)
    set(CMAKE_C_FLAGS -mx32)
    set(CMAKE_CXX_FLAGS -mx32)
else()
    message(FATAL_ERROR "Unsupported package arch!")
endif ()

message(PACKAGE_VERSION="$ENV{PACKAGE_VERSION}")
message(PACKAGE_ARCH="$ENV{PACKAGE_ARCH}")
message(PLATFORM_OS="$ENV{PLATFORM_OS}")
message(PLATFORM_ARCH="$ENV{PLATFORM_ARCH}")
message(DSE_MODELC_LIB_URL="$ENV{DSE_MODELC_LIB_URL}")

project(Models
    VERSION $ENV{PACKAGE_VERSION}
    DESCRIPTION "DSE DSP Example Models"
)
include(GNUInstallDirs)
include(FetchContent)
set(CDEF_PLATFORM_OS $ENV{PLATFORM_OS})
set(CDEF_PLATFORM_ARCH $ENV{PLATFORM_ARCH})
set(CMAKE_INSTALL_PREFIX ${CMAKE_BINARY_DIR}/_out)
set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED TRUE)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_C_FLAGS_DEBUG "-g -ggdb")
set(CMAKE_C_FLAGS_DEBUG "${CMAKE_C_FLAGS_DEBUG} -O3")
list(APPEND C_CXX_WARNING_FLAGS
    -Wall
    -W
    -Wwrite-strings
    -Wno-missing-field-initializers
    -Wno-misleading-indentation
)
add_compile_options(${C_CXX_WARNING_FLAGS})
add_compile_definitions(DLL_BUILD)

FetchContent_Declare(dse_modelc_lib URL $ENV{DSE_MODELC_LIB_URL})
FetchContent_MakeAvailable(dse_modelc_lib)

add_subdirectory(simple)
add_subdirectory(gateway)
add_subdirectory(network_pdu)
add_subdirectory(network_can)

set(CPACK_SYSTEM_NAME $ENV{PACKAGE_ARCH})
set(CPACK_PACKAGE_VENDOR "Robert Bosch GmbH")
set(CPACK_PACKAGE_DESCRIPTION "SDP Example Models")
set(CPACK_OUTPUT_FILE_PREFIX _dist)
set(CPACK_GENERATOR "ZIP")
set(CPACK_MONOLITHIC_INSTALL TRUE)
include(CPack)

