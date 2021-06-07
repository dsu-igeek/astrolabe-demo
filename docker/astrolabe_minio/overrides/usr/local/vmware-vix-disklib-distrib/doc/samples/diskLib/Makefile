INCLUDEDIR=../../../include
LIBDIR=../../../lib64
LIBS=-ldl -lz -lsqlite3 -lcurl -lssl -lcrypto -lcares

ifdef VIX_AIO_BUFPOOL_SIZE
CXXFLAGS+= -DVIX_AIO_BUFPOOL_SIZE=$(VIX_AIO_BUFPOOL_SIZE)
endif

CXXFLAGS+= -std=c++1y -lpthread

all: vix-disklib-sample vix-mntapi-sample

vix-disklib-sample: vixDiskLibSample.cpp
	$(CXX) $(CXXFLAGS) -o $@ -I$(INCLUDEDIR) -L$(LIBDIR) $? $(LIBS) -lvixDiskLib

vix-mntapi-sample:  vixDiskLibSample.cpp
	$(CXX) $(CXXFLAGS) -o $@ -DFOR_MNTAPI -I$(INCLUDEDIR) -L$(LIBDIR) $? $(LIBS) \
	   -lfuse -lvixDiskLib -lvixMntapi

clean:
	$(RM) -f vix-disklib-sample vix-mntapi-sample

