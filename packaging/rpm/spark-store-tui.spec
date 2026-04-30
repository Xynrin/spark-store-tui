Name:           spark-store-tui
Version:        0.7.2
Release:        1%{?dist}
Summary:        Terminal UI for browsing Spark Store and APM Store

License:        GPL-3.0-only
URL:            https://github.com/Xynrin/spark-store-tui
Source0:        https://github.com/Xynrin/%{name}/releases/download/v%{version}/%{name}-deb-source-%{version}.tar.gz

BuildArch:      noarch
Requires:       bash
Requires:       curl
Requires:       jq
Requires:       fzf
Requires:       aria2
Requires:       ca-certificates
Recommends:     chafa
Recommends:     sudo

%description
Spark Store TUI is a Bash and fzf based terminal interface for browsing Spark
Store and APM Store metadata, previewing application information, and
downloading packages through Metalink and aria2c.

%prep
%autosetup -n %{name}-deb-source-%{version}

%build

%install
install -Dm0755 package-root/usr/bin/%{name} %{buildroot}%{_bindir}/%{name}
install -Dm0644 README.md %{buildroot}%{_docdir}/%{name}/README.md
install -Dm0644 COPYING %{buildroot}%{_licensedir}/%{name}/COPYING

%files
%license %{_licensedir}/%{name}/COPYING
%doc %{_docdir}/%{name}/README.md
%{_bindir}/%{name}

%changelog
* Thu Apr 30 2026 Xynrin <xynrin@163.com> - 0.7.2-1
- Initial RPM package
