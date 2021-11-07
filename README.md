<h1 align="center">
    Sasm-docker
</h1>

<p align="center">
    <!--a href="https://www.gnu.org/licenses/agpl-3.0">
        <img src="https://img.shields.io/badge/License-AGPL%20v3-blue.svg" />
    </a-->
    <a href="https://github.com/keinenclue/sasm-docker/releases/latest">
        <img src="https://img.shields.io/github/v/release/keinenclue/sasm-docker?logo=github&logoColor=white" alt="GitHub release"/>
    </a>
    <a href="https://github.com/keinenclue/sasm-docker/actions/workflows/release.yml/badge.svg">
        <img src="https://github.com/keinenclue/sasm-docker/actions/workflows/release.yml/badge.svg" alt="Badge tests">
    </a>
    <a href="https://goreportcard.com/report/github.com/keinenclue/sasm-docker">
        <img src="https://img.shields.io/badge/go%20report-A-green.svg?style=flat" alt="Go report" />
    </a>
</p>

Sasm-docker simplifies the setup and use of [SASM](https://dman95.github.io/SASM/english.html) by running it inside a docker container and using x11 (X Window System) in order to display the SASM GUI.

## Features

- **Easy setup:** Just install docker and xserver, download the launcher, and you're ready to go
- **Easy updating:** The launcher checks for and downloads updates on every start

## Run
#### Install X server
- On MacOS [follow these steps](https://gist.github.com/paul-krohn/e45f96181b1cf5e536325d1bdee6c949) but use `xhost +localhost` instead of `xhost +$(hostname).local`.
- On Windows ...
- On Linux you are probaply ready to go :)

#### Install docker
- On MaxOS ...
- On Windows ...
- On Linux [follow these steps](https://docs.docker.com/engine/install)

#### Install the launcher
-  Download the binary for your OS over here: [https://github.com/keinenclue/sasm-docker/releases/latest](https://github.com/keinenclue/sasm-docker/releases/latest)

## Screenshots
<table align="center">
    <tr>
        <td align="center">
               <a href="https://user-images.githubusercontent.com/30153207/140638832-c3f91a51-81a3-44db-8a1e-0192c9fe9ec5.png">
                   <img src="https://user-images.githubusercontent.com/30153207/140638832-c3f91a51-81a3-44db-8a1e-0192c9fe9ec5.png" width="500px" alt="Screenshot launch" />
              </a>
        </td>
        <td align="center">
            <a href="https://user-images.githubusercontent.com/30153207/140639058-fed938e4-2a66-4a42-849d-86c5a4fb61a6.png" >
                <img src="https://user-images.githubusercontent.com/30153207/140639058-fed938e4-2a66-4a42-849d-86c5a4fb61a6.png" width="500px" alt="Screenshot pulling" />
            </a>
        </td>
    </tr>
    <tr>
       <td align="center">
            <a href="https://user-images.githubusercontent.com/30153207/140638903-1e6cde6f-e561-4d52-901f-2c4cf784fae6.png">
                <img src="https://user-images.githubusercontent.com/30153207/140638903-1e6cde6f-e561-4d52-901f-2c4cf784fae6.png"  width="500px" alt="Screenshot logs" />
            </a>
        </td>
        <td align="center">
            <a href="https://user-images.githubusercontent.com/30153207/140639009-8f6dd323-12aa-4e8f-9d6d-afbcfed45e32.png" >
                <img src="https://user-images.githubusercontent.com/30153207/140639009-8f6dd323-12aa-4e8f-9d6d-afbcfed45e32.png" width="500px" alt="Screenshot settings" />
            </a>
        </td>
    </tr>
</table>

