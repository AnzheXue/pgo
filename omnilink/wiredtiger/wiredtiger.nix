{
  stdenv,
  python3,
  swig,
  cmake,
  pkg-config,
  fetchFromGitHub,
  omnilink,
  msgpack-cxx,

  ghUrl ? "https://github.com/wiredtiger/wiredtiger.git",
  ghRev,
}:
let
  src = (builtins.fetchGit {
    url = ghUrl;
    rev = ghRev;
  }).outPath;
in
stdenv.mkDerivation {
  version = ghRev;
  pname = "wiredtiger";
  dontStrip = true;
  env.NIX_CFLAGS_COMPILE = "-Wno-error"; # or we get strange fireworks!
  cmakeFlags = [
    # unbreak .pc file
    "-DCMAKE_INSTALL_LIBDIR=lib"
    "-DCMAKE_INSTALL_INCLUDEDIR=include"
    # debug build in case something goes wrong
    "-DCMAKE_BUILD_TYPE=Debug"
    # allows us to extract wtperf
    "-DENABLE_STATIC=ON"
    "-DENABLE_SHARED=OFF"
    # This one's broken for some reason.
    "-DENABLE_PYTHON=OFF"
  ];
  buildInputs = [
    python3
    swig
    omnilink.lib
    msgpack-cxx
    omnilink.wiredtiger.reflocking_wrapper
  ];
  nativeBuildInputs = [
    cmake
    pkg-config
  ];
  env.WIREDTIGER_SRC = src;
  unpackPhase = ''
    echo skipping unpack
  '';
  configurePhase = ''
    mkdir build
    cmake -B ./build -S "$WIREDTIGER_SRC" $cmakeFlags
  '';
  buildPhase = ''
    cmake --build ./build
  '';
  installPhase = ''
    cmake --install ./build --prefix $out
  '';
  postInstall = ''
    cp build/bench/wtperf/wtperf $out/bin/wtperf
  '';
}