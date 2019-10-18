
apply_win64() {
  export AS=/usr/src/mxe/usr/bin/${CROSS_TRIPLE}-as
  export AR=/usr/src/mxe/usr/bin/${CROSS_TRIPLE}-ar
  export CC=/usr/src/mxe/usr/bin/${CROSS_TRIPLE}-gcc
  export CPP=/usr/src/mxe/usr/bin/${CROSS_TRIPLE}-cpp
  export CXX=/usr/src/mxe/usr/bin/${CROSS_TRIPLE}-g++
  export LD=/usr/src/mxe/usr/bin/${CROSS_TRIPLE}-ld
  export FC=/usr/src/mxe/usr/bin/${CROSS_TRIPLE}-gfortran
  export CMAKE_TOOLCHAIN_FILE=/usr/src/mxe/usr/${CROSS_TRIPLE}/share/cmake/mxe-conf.cmake
  export GOOS=windows
  export GW_BUILD_TARGET=/${GOOS}

  cd /usr/bin
  rm cmake cpack
  ln -s /usr/src/mxe/usr/bin/${CROSS_TRIPLE}-cmake cmake
  ln -s /usr/src/mxe/usr/bin/${CROSS_TRIPLE}-cpack cpack
  cd -
}

apply_osx64() {
  export AS=${OSX_CROSS_PATH}/target/bin/x86_64-apple-darwin14-as
  export AR=${OSX_CROSS_PATH}/target/bin/x86_64-apple-darwin14-ar
  export CC=${OSX_CROSS_PATH}/target/bin/x86_64-apple-darwin14-clang
  export CPP=${OSX_CROSS_PATH}/target/bin/x86_64-apple-darwin14-clang++
  export CXX=${OSX_CROSS_PATH}/target/bin/x86_64-apple-darwin14-clang++
  export LD=${OSX_CROSS_PATH}/target/bin/x86_64-apple-darwin14-ld
  export FC=
  export CMAKE_TOOLCHAIN_FILE=/usr/src/mxe/usr/${CROSS_TRIPLE}/share/cmake/mxe-conf.cmake
  export GOOS=darwin
  export GW_BUILD_TARGET=/${GOOS}

  cd /usr/bin
  rm cmake cpack
  ln -s /usr/src/cmake-3.13.2-Centos5-x86_64/bin/cmake cmake
  ln -s /usr/src/cmake-3.13.2-Centos5-x86_64/bin/cpack cpack
  cd -
}

restore_linux() {
  export AS=/usr/bin/as
  export AR=/usr/bin/ar
  export CC=/usr/bin/gcc
  export CPP=/usr/bin/cpp
  export CXX=/usr/bin/g++
  export LD=/usr/bin/ld
  unset CMAKE_TOOLCHAIN_FILE
  export GOOS=linux
  export GW_BUILD_TARGET=/${GOOS}

  cd /usr/bin
  rm cmake cpack
  ln -s /usr/src/cmake-3.13.2-Centos5-x86_64/bin/cmake cmake
  ln -s /usr/src/cmake-3.13.2-Centos5-x86_64/bin/cpack cpack
  cd -
}

restore_linux_clang() {
  export AS=/usr/bin/as
  export AR=/usr/bin/ar
  export CC=/usr/bin/clang
  export CPP=/usr/bin/clang++
  export CXX=/usr/bin/clang++
  export LD=/usr/bin/ld
  unset CMAKE_TOOLCHAIN_FILE
  export GOOS=linux
  export GW_BUILD_TARGET=/${GOOS}

  cd /usr/bin
  rm cmake cpack
  ln -s /usr/src/cmake-3.13.2-Centos5-x86_64/bin/cmake cmake
  ln -s /usr/src/cmake-3.13.2-Centos5-x86_64/bin/cpack cpack
  cd -
}
